# EduSpark

## About
AI enhanced application can help teachers to teach.
---

## **EduSpark Git 提交规范**
#### **1. 克隆仓库并切换到 `dev` 分支**
1. 使用 SSH 克隆远程仓库：
   ```bash
   git clone git@github.com:hedgeho9X/EduSpark.git
   ```

2. 切换到 `dev` 分支：
   ```bash
   git checkout dev
   ```

3. 确保 `dev` 分支是最新的：
   ```bash
   git pull origin dev
   ```

---

#### **2. 开发工作流程**
所有开发必须基于 `dev` 分支，按照以下步骤提交和推送代码：

1. 确保开始开发前，拉取最新的 `dev` 分支代码：
   ```bash
   git pull origin dev
   ```

2. 开发过程中，完成代码后将更改添加到暂存区：
   ```bash
   git add .
   ```

3. 提交代码时，请使用规范的提交信息格式：
   ```bash
   git commit -m "类别: 描述本次提交内容"
   ```
   ##### 提交信息分类：
   - `feat`: 新功能
   - `fix`: 修复问题
   - `docs`: 文档更新
   - `style`: 格式修改（不影响代码逻辑）
   - `refactor`: 重构代码
   - `test`: 添加测试

   示例：
   ```bash
   git commit -m "feat: 添加用户登录功能"
   ```

4. 提交完成后，将 `dev` 分支推送到远程仓库：
   ```bash
   git push origin dev
   ```
---

#### **4. 提交规范总结**
##### 示例开发流程：
```bash
# 克隆仓库
git clone git@github.com:hedgeho9X/DouMall.git

# 切换到 dev 分支
git checkout dev

# 拉取最新代码
git pull origin dev

# 开发代码
# ...（代码修改）

# 添加到暂存区
git add .

# 提交代码（遵循提交信息规范）
git commit -m "feat: 添加商品搜索功能"

# 推送到远程 dev 分支
git push origin dev
```

#### 常用 Git 命令：
- 查看当前分支：
  ```bash
  git branch
  ```
- 查看提交记录：
  ```bash
  git log --oneline
  ```
- 拉取最新代码：
  ```bash
  git pull origin dev
  ```
- 推送代码：
  ```bash
  git push origin dev
  ```
