# 在文件最开头添加
__import__('pysqlite3')
import sys
sys.modules['sqlite3'] = sys.modules.pop('pysqlite3')
import os
import torch
from typing import List, Any
from transformers import AutoModel, AutoTokenizer
from pydantic import BaseModel
from langchain_core.embeddings import Embeddings
from langchain_community.document_loaders import (
    DirectoryLoader,
    UnstructuredMarkdownLoader,
    TextLoader
)
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_chroma import Chroma
from langchain_openai import ChatOpenAI
from langchain.chains import RetrievalQA
from langchain.prompts import PromptTemplate
from modelscope import snapshot_download

# ==================== 配置参数 ====================
MD_DIR = "/root"  # 文档目录路径
CHROMA_DIR = "/root/chroma_db" # ChromaDB持久化目录
MODEL_CACHE_DIR = "/root/model_cache" # 模型缓存目录
BATCH_SIZE = 64      # GPU批处理大小
CHUNK_SIZE = 800    # 文本块大小
CHUNK_OVERLAP = 250 # 块重叠量
DEVICE = "mps" if torch.backends.mps.is_available() else "cpu" # 自动选择设备 (优先 MPS)

# ==================== Embedding模型定义 ====================
class BgeM3Embeddings(BaseModel, Embeddings):
    """
    基于 BAAI/bge-m3 模型的自定义嵌入类。

    该类负责下载模型、初始化 tokenizer 和模型，并提供文本嵌入功能。
    它继承自 pydantic.BaseModel 用于数据验证，并实现 langchain_core.embeddings.Embeddings 接口。

    Attributes:
        model_name (str): 要使用的预训练模型的名称。
        tokenizer (Any): 用于文本编码的分词器实例。
        model (Any): 加载的预训练嵌入模型实例。
    """
    model_name: str = "BAAI/bge-m3"
    tokenizer: Any = None
    model: Any = None

    def __init__(self, **data: Any):
        """
        初始化 BgeM3Embeddings 类。

        下载模型（如果尚未缓存），并加载 tokenizer 和模型到指定设备。

        Args:
            **data (Any): Pydantic BaseModel 的初始化参数。
        """
        super().__init__(**data)
        os.makedirs(MODEL_CACHE_DIR, exist_ok=True)
        # 使用魔搭社区下载模型
        model_path = snapshot_download(self.model_name,
                                     cache_dir=MODEL_CACHE_DIR,
                                     revision='master')

        self.tokenizer = AutoTokenizer.from_pretrained(model_path)
        self.model = AutoModel.from_pretrained(
            model_path,
            torch_dtype=torch.float16 if torch.backends.mps.is_available() else torch.float32, # 优先使用 float16 以节省 MPS 内存
            offload_folder=MODEL_CACHE_DIR # 指定模型参数卸载目录
        ).eval().to(DEVICE) # 将模型设置为评估模式并移动到目标设备

    def _embed(self, text: str) -> List[float]:
        """
        生成单个文本字符串的嵌入向量。

        Args:
            text (str): 需要嵌入的文本。

        Returns:
            List[float]: 表示文本的嵌入向量。
        """
        inputs = self.tokenizer(
            text,
            padding=True,
            truncation=True,
            return_tensors="pt",
            max_length=512 # 根据模型设定最大长度
        ).to(DEVICE)

        with torch.no_grad(): # 禁用梯度计算以节省内存和加速
            outputs = self.model(**inputs)

        # 提取 [CLS] token 的隐藏状态作为句子嵌入
        return outputs.last_hidden_state[:, 0].cpu().numpy().flatten().tolist()

    def embed_query(self, text: str) -> List[float]:
        """
        为单个查询文本生成嵌入向量。 Langchain Embeddings 接口要求的方法。

        Args:
            text (str): 查询文本。

        Returns:
            List[float]: 查询文本的嵌入向量。
        """
        return self._embed(text)

    def embed_documents(self, texts: List[str]) -> List[List[float]]:
        """
        为文档列表批量生成嵌入向量。

        Args:
            texts (List[str]): 需要嵌入的文档文本列表。

        Returns:
            List[List[float]]: 每个文档对应的嵌入向量列表。
        """
        all_embeddings = []
        for i in range(0, len(texts), BATCH_SIZE):
            batch = texts[i:i + BATCH_SIZE]
            inputs = self.tokenizer(
                batch,
                padding=True,
                truncation=True,
                return_tensors="pt",
                max_length=512
            ).to(DEVICE)

            # 使用自动混合精度（AMP）进行推理，如果设备支持（例如 CUDA）
            # 注意：MPS 对 AMP 的支持可能有限或不同，这里保留 'cuda' 以便兼容性
            # 如果在 MPS 上遇到问题，可以移除 torch.amp.autocast
            with torch.no_grad(), torch.amp.autocast(device_type='cuda' if DEVICE == 'cuda' else 'cpu'):
                outputs = self.model(**inputs)

            # 提取 [CLS] token 的嵌入并转换为 NumPy 数组
            embeddings = outputs.last_hidden_state[:, 0].cpu().numpy()
            # 转换为列表的列表结构
            all_embeddings.extend([emb.tolist() for emb in embeddings])

            # 打印进度信息
            if (i + BATCH_SIZE) % 1000 < BATCH_SIZE and i != 0: # 避免在每个 batch 都打印
                 print(f"已处理 {min(i + BATCH_SIZE, len(texts))}/{len(texts)} 个文档")

        return all_embeddings

# ==================== 全局初始化 ====================
embeddings = BgeM3Embeddings() # 初始化嵌入模型实例
llm = ChatOpenAI(
    model=os.getenv("LLM_MODELEND"), # 从环境变量获取 LLM 模型名称
    temperature=0 # 设置温度为 0 以获得更确定的输出
)

def init_vectorstore() -> Chroma:
    """
    初始化或加载 Chroma 向量数据库。

    如果 CHROMA_DIR 存在，则加载现有数据库；否则，从 MD_DIR 中的 Markdown 文件
    创建新的向量数据库，并将文档分割、嵌入后存入。

    Returns:
        Chroma: 初始化或加载的 Chroma 向量数据库实例。
    """
    if os.path.exists(CHROMA_DIR):
        print(f">> 加载已有向量库: {CHROMA_DIR}")
        # 加载持久化的向量数据库
        return Chroma(persist_directory=CHROMA_DIR, embedding_function=embeddings)

    print(">> 创建新向量库...")
    # 配置文档加载器，加载指定目录下的所有 Markdown 文件
    loader = DirectoryLoader(MD_DIR, glob="**/*.md",
                           loader_cls=UnstructuredMarkdownLoader, show_progress=True)

    docs = loader.load()
    print(f"原始文档数: {len(docs)}")

    # 配置文本分割器
    # 使用正则表达式匹配更广泛的题号格式（阿拉伯数字和中文数字）
    # r"\n\d+\.\s+" 匹配如 "\n1. ", "\n10. " 等
    # r"\n[一二三四五六七八九十百千万]+\s?[.．、]\s?" 匹配如 "\n一. ", "\n十、", "\n二十三．" 等
    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=CHUNK_SIZE,
        chunk_overlap=CHUNK_OVERLAP,
        # 将正则表达式模式放在前面，优先尝试按题号分割
        # is_separator_regex=True, # Langchain v0.2+ 可能需要此参数，旧版本可能不需要
        separators=[
            r"\n\d+\.\s+",                             # 匹配阿拉伯数字题号 (如: \n1. , \n23. )
            r"\n[一二三四五六七八九十百千万]+\s?[.．、]\s?", # 匹配中文数字题号 (如: \n一. , \n十、 , \n二十三．)
            # 保留原来的 Markdown 结构分隔符作为后备
            "\n# ", "\n## ", "\n### ", "\n\n", "\n", " "
            ]
    )

    texts = text_splitter.split_documents(docs)
    print(f"分割后文本块数: {len(texts)}")

    # 分批处理并创建向量库，以防文档过多导致内存问题
    vectorstore = None
    batch_size = 5000  # 每批处理的文档块数

    for i in range(0, len(texts), batch_size):
        batch = texts[i:i + batch_size]
        print(f"\n处理批次 {i//batch_size + 1}/{(len(texts) + batch_size - 1)//batch_size}")

        if vectorstore is None:
            # 创建第一个批次的向量库
            vectorstore = Chroma.from_documents(
                documents=batch,
                embedding=embeddings,
                persist_directory=CHROMA_DIR,
                collection_metadata={"hnsw:space": "cosine"} # 指定使用余弦相似度
            )
        else:
            # 将后续批次添加到现有向量库
            vectorstore.add_documents(batch)

        vectorstore.persist()  # 每处理完一个批次后持久化保存
        print(f"已处理并保存 {min(i + batch_size, len(texts))}/{len(texts)} 个文档块")

    return vectorstore

# ==================== 全局初始化 ====================
# 定义用于问答链的 Prompt 模板
PROMPT_TEMPLATE = """根据以下上下文推荐相似试题：
{context}

问题：{question}
包括选择,填空,解答等题型
推荐试题及答案：
"""

# ==================== 查询接口 ====================
class QAEngine:
    """
    问答引擎类，封装了向量检索和语言模型调用。

    该类初始化向量数据库检索器和问答链，并提供查询接口。
    """
    def __init__(self):
        """
        初始化 QAEngine。

        加载或创建向量数据库，配置检索器和问答链。
        """
        # 初始化向量数据库
        self.vectorstore = init_vectorstore()

        # 配置检索器，使用相似度阈值过滤结果
        self.retriever = self.vectorstore.as_retriever(
            search_type="similarity_score_threshold",
            search_kwargs={
                "k": 8,            # 返回最多 8 个结果
                "score_threshold": 0.5  # 相似度得分阈值
            }
        )

        # 创建自定义 Prompt 模板
        custom_prompt = PromptTemplate(
            template=PROMPT_TEMPLATE,
            input_variables=["context", "question"]
        )

        # 初始化 RetrievalQA 链
        self.chain = RetrievalQA.from_chain_type(
            llm=llm,
            chain_type="stuff", # 使用 "stuff" 链类型，将所有检索到的文档放入 Prompt
            retriever=self.retriever,
            return_source_documents=True, # 返回源文档信息
            chain_type_kwargs={"prompt": custom_prompt} # 应用自定义 Prompt
        )

    def query(self, question: str) -> str:
        """
        执行问答查询并返回格式化的结果字符串。

        Args:
            question (str): 用户提出的问题。

        Returns:
            str: 包含答案和参考资料来源的格式化字符串。如果发生错误，则返回错误信息。
        """
        try:
            result = self.chain.invoke({"query": question})
            answer = result['result'].strip() # 提取并清理答案文本

            # 格式化源文档信息
            sources = "\n".join(
                f"- {doc.metadata.get('source', '未知来源')}" # 获取每个源文档的文件路径
                for doc in result['source_documents']
            )
            return f"{answer}\n\n参考资料:\n{sources}"
        except Exception as e:
            # 捕获并返回查询过程中可能出现的任何异常
            return f"查询错误: {str(e)}"

    def query_with_metadata(self, question: str) -> dict:
        """
        执行问答查询并返回包含详细元数据的字典。

        Args:
            question (str): 用户提出的问题。

        Returns:
            dict: 包含答案和详细源文档信息（内容和元数据）的字典。如果发生错误，则返回包含错误信息的字典。
        """
        try:
            result = self.chain.invoke({"query": question})
            return {
                "answer": result['result'].strip(),
                "sources": [
                    {
                        "content": doc.page_content, # 源文档块内容
                        "metadata": doc.metadata     # 源文档元数据
                    } for doc in result['source_documents']
                ]
            }
        except Exception as e:
            # 捕获并返回查询过程中可能出现的任何异常
            return {"error": str(e)}

# ==================== 主程序 ====================
if __name__ == "__main__":
    """
    主程序入口点。

    初始化问答引擎并启动一个循环，接收用户输入并打印查询结果，直到用户输入 'q' 退出。
    """
    # 初始化问答引擎
    qa_engine = QAEngine()

    # 启动交互式查询循环
    while True:
        question = input("\n请输入问题（输入q退出）: ")
        if question.lower() == 'q':
            break
        # 执行查询并打印结果
        print("\n", qa_engine.query(question))