# Table-drives-LL-1-parser

任务要求为：
根据LL(1)分析法编写一个语法分析程序，输入已知文法，判断该文法是否为LL(1)文法，由程序自动构造文法的预测分析表。对输入的任意符号串，所编制的语法分析程序应能正确判断此串是否为文法的句子。

结构设计为：  
type Symbol struct {
	Value      string
	IsTerminal bool
}  
type Alternative struct {
	Symbols []Symbol
}  
type Production struct {
	Left  Symbol
	Right []Alternative
}  
type Grammar struct {
	Start       Symbol
	Productions []Production
	Nullable    map[string]bool
	FirstSet    map[Symbol]map[Symbol]bool
	FollowSet   map[Symbol]map[Symbol]bool
	Predict     map[Symbol]map[Symbol]Production
}  

主要提供的方法：
func (g *Grammar) GetNonTerminals() []Symbol  
func (g *Grammar) GetTerminals() []Symbol  
func (g *Grammar) AllNullable(symbols []Symbol) bool  
func (g Grammar) GetFirst(symbols []Symbol) []Symbol  
func (g Grammar) Select(left Symbol, right []Symbol) []Symbol  
func (g Grammar) parse(strs string)  
func printStep(w *tabwriter.Writer, step int, analysisStack []Symbol, characterStack []string)   
func (g *Grammar) extractCommonFactors()  
func findLongestCommonPrefix(alternatives []Alternative) []Symbol  
func removeCommonPrefix(alternatives []Alternative, prefix []Symbol) []Alternative  
func (g *Grammar) GInit() bool   
func (g *Grammar) initializeNullable()  
func (g *Grammar) initializeFirstSet()  
func (g *Grammar) initializeFollowSet()  
func (g *Grammar) initializePredict()  
func findString(strs []Symbol, s string) bool  
func (g Grammar) isLL1()  
func hasIntersection(slice1, slice2 []Symbol) bool  
func symbolsToString(symbols []Symbol) string  
func (g *Grammar) PrintGrammar()  
func (g *Grammar) PrintNonTerminals()  
func (g *Grammar) PrintTerminals()  
func printSymbolSlice(symbols []Symbol)  
func (g *Grammar) PrintNullableTable()  
func (g *Grammar) PrintFirstSet()  
func (g *Grammar) PrintFollowSet()   
func (g *Grammar) PrintPredict()  
func deleteSymbol(symbols []Symbol, value string) []Symbol  
func addSymbol(symbols []Symbol, value string, isTerminal bool) []Symbol  
  
其中代码严格按照代码规范编写，首字母大写的函数可供调用，创建一个Grammar之后必须调用GInit完成初始化




