# Table-drives-LL-1-parser

# 任务要求
根据LL(1)分析法编写一个语法分析程序，输入已知文法，判断该文法是否为LL(1)文法，由程序自动构造文法的预测分析表。对输入的任意符号串，所编制的语法分析程序应能正确判断此串是否为文法的句子。

# 分析过程
1. 处理文法
2. 对输入的文法进行分析，提取终结符集和非终结符集
3. 计算first集，follow集，select集
4. 判断是否为LL（1）文法
5. 生成预测分析表
6. 对输入的句子进行分析

# 测试用例
#### LL（1）文法：
+ S->AaS|BbS|d
+ A->a  
+ B->ε|c

#### 非LL（1）文法：
+ S->A|B
+ A->Aab|Aac|cd|e
+ B->b|e




