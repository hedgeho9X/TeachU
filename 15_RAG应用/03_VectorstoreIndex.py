""" 
本文件是【检索增强生成：通过 RAG 助力鲜花运营】章节的配套代码，课程链接：https://juejin.cn/book/7387702347436130304/section/7388069959185727524
您可以点击最上方的“运行“按钮，直接运行该文件；更多操作指引请参考Readme.md文件。
"""
# 设置OpenAI的API密钥
import os
from volcenginesdkarkruntime import Ark
from typing import List, Any
from langchain.embeddings.base import Embeddings
from langchain.pydantic_v1 import BaseModel
from langchain.document_loaders import UnstructuredWordDocumentLoader

# 初始化Embedding类
class DoubaoEmbeddings(BaseModel, Embeddings):
    client: Ark = None
    api_key: str = ""
    model: str

    def __init__(self, **data: Any):
        super().__init__(**data)
        if self.api_key == "":
            self.api_key = os.environ["OPENAI_API_KEY"]
        self.client = Ark(
            base_url=os.environ["OPENAI_BASE_URL"],
            api_key=self.api_key
        )

    def embed_query(self, text: str) -> List[float]:
        """
        生成输入文本的 embedding.
        Args:
            texts (str): 要生成 embedding 的文本.
        Return:
            embeddings (List[float]): 输入文本的 embedding，一个浮点数值列表.
        """
        embeddings = self.client.embeddings.create(model=self.model, input=text)
        return embeddings.data[0].embedding

    def embed_documents(self, texts: List[str]) -> List[List[float]]:
        return [self.embed_query(text) for text in texts]

    class Config:
        arbitrary_types_allowed = True


embeddings = DoubaoEmbeddings(
    model=os.environ["EMBEDDING_MODELEND"],
)

# 导入文档加载器模块，并使用TextLoader来加载文本文件
from langchain_community.document_loaders import TextLoader
from langchain_community.document_loaders import UnstructuredWordDocumentLoader
from langchain_openai import ChatOpenAI  # ChatOpenAI模型

#loader = TextLoader("./OneFlower/花语大全.txt", encoding="utf8")
loader = UnstructuredWordDocumentLoader("./OneFlower/易速鲜花运营指南.docx")

# 使用VectorstoreIndexCreator来从加载器创建索引
from langchain.indexes import VectorstoreIndexCreator

index = VectorstoreIndexCreator(embedding=embeddings).from_loaders([loader])

llm = ChatOpenAI(model=os.environ["LLM_MODELEND"], temperature=0)

# 定义查询字符串, 使用创建的索引执行查询
query = "rose的花语"
result = index.query(llm=llm, question=query)
print(result)  # 打印查询结果

# 替换成你所需要的工具
from langchain.text_splitter import CharacterTextSplitter

text_splitter = CharacterTextSplitter(chunk_size=1000, chunk_overlap=0)
from langchain_community.vectorstores import Qdrant

index_creator = VectorstoreIndexCreator(
    vectorstore_cls=Qdrant,
    embedding=embeddings,
    text_splitter=CharacterTextSplitter(chunk_size=1000, chunk_overlap=0),
)
