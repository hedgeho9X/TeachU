"""
本文件是【检索增强生成：通过 RAG 助力鲜花运营】章节的配套代码
课程链接：https://juejin.cn/book/7387702347436130304/section/7388069959185727524
您可以点击最上方的“运行“按钮，直接运行该文件；更多操作指引请参考Readme.md文件。
"""

# ------------ 环境配置 ------------
# 设置OpenAI的API密钥相关环境变量
import os  # 用于读取系统环境变量
from volcenginesdkarkruntime import Ark  # 火山引擎的SDK
from typing import List, Any  # 类型提示支持
from langchain.embeddings.base import Embeddings  # LangChain的Embedding基类
from langchain.pydantic_v1 import BaseModel  # 数据验证基类
from langchain.document_loaders import UnstructuredWordDocumentLoader  # Word文档加载器

# ------------ 自定义Embedding类 ------------
class DoubaoEmbeddings(BaseModel, Embeddings):
    """
    自定义的火山引擎（豆包）Embedding实现类
    继承自BaseModel（数据验证）和Embeddings（LangChain接口）
    """
    client: Ark = None  # 火山引擎客户端实例
    api_key: str = ""  # API密钥
    model: str  # 使用的模型名称

    def __init__(self, **data: Any):
        """初始化方法，自动从环境变量获取API配置"""
        super().__init__(**data)
        # 如果未显式设置api_key，则从环境变量获取
        if self.api_key == "":
            self.api_key = os.environ["OPENAI_API_KEY"]  # 实际使用火山引擎的API_KEY
        # 初始化火山引擎客户端
        self.client = Ark(
            base_url=os.environ["OPENAI_BASE_URL"],  # 服务端地址
            api_key=self.api_key  # 身份验证密钥
        )

    def embed_query(self, text: str) -> List[float]:
        """
        生成单个文本的embedding向量
        Args:
            text (str): 输入文本，长度建议不超过256个token
        Returns:
            List[float]: 384/768维的浮点数向量（具体维度取决于模型）
        """
        embeddings = self.client.embeddings.create(
            model=self.model,  # 使用的embedding模型
            input=text         # 需要编码的文本
        )
        return embeddings.data[0].embedding  # 从响应中提取向量

    def embed_documents(self, texts: List[str]) -> List[List[float]]:
        """批量生成文本embedding，直接循环调用embed_query"""
        return [self.embed_query(text) for text in texts]

    class Config:
        arbitrary_types_allowed = True  # 允许非Pydantic类型字段

# 初始化embedding实例（实际使用火山引擎的模型）
embeddings = DoubaoEmbeddings(
    model=os.environ["EMBEDDING_MODELEND"],  # 从环境变量获取模型名称
)

# ------------ 文档加载 ------------
from langchain_community.document_loaders import UnstructuredWordDocumentLoader
from unstructured.partition.auto import partition
from unstructured.partition.text import partition_text

# 修正后的文档加载配置
loader = UnstructuredWordDocumentLoader(
    "函数.docx",
    mode="elements",  # 使用elements模式以获取更好的结构化内容
    strategy="fast"   # 使用快速解析策略
)

# 打印加载的文档内容（保留公式结构）
documents = loader.load()
for i, doc in enumerate(documents):
    print(f"文档段落 {i+1}:")
    print("-" * 50)
    # 保留换行符并显示公式标记
    print(doc.page_content.replace('\n', '\n↲ '))  # ↲ 表示换行符
    print("-" * 50)
    print("元数据：", doc.metadata)
    print("\n")

# ------------ 索引创建 ------------
from langchain.indexes import VectorstoreIndexCreator

# 创建向量存储索引（自动处理文本分块和向量化）
index = VectorstoreIndexCreator(embedding=embeddings).from_loaders([loader])

# ------------ 大语言模型配置 ------------
llm = ChatOpenAI(
    model=os.environ["LLM_MODELEND"],  # 从环境变量获取大模型名称
    temperature=0  # 控制生成结果的随机性（0表示最确定性输出）
)

# ------------ 查询执行 ------------
query = "公切线问题咋做"  # 用户提问
result = index.query(
    llm=llm,     # 使用配置的大模型
    question=query  # 要回答的问题
)
print(result)  # 输出：玫瑰的常见花语是...

# ------------ 备用配置（当前未使用） ------------
from langchain.text_splitter import CharacterTextSplitter  # 文本分块工具
from langchain_community.vectorstores import Qdrant  # 向量数据库

# 文本分块配置（用于自定义索引创建）
text_splitter = CharacterTextSplitter(
    chunk_size=1000,   # 每个文本块的最大长度
    chunk_overlap=0    # 块之间的重叠长度
)

# 备用索引创建器配置（使用Qdrant向量数据库）
index_creator = VectorstoreIndexCreator(
    vectorstore_cls=Qdrant,  # 指定向量数据库类型
    embedding=embeddings,    # 使用相同的embedding模型
    text_splitter=CharacterTextSplitter(
        chunk_size=1000, 
        chunk_overlap=0
    )
)

# 创建向量存储索引（使用持久化存储）
# index = VectorstoreIndexCreator(
#     vectorstore_cls=Qdrant,  # 使用 Qdrant 作为持久化向量存储
#     embedding=embeddings,
#     text_splitter=CharacterTextSplitter(chunk_size=1000, chunk_overlap=0)
# ).from_loaders([loader])
