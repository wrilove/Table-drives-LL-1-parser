package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

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

// GetNonTerminals
//获取输出的文法所有产生式的非终结符
func (g *Grammar) GetNonTerminals() []Symbol {
	nonTerminals := make(map[string]bool)

	// 遍历所有产生式规则，将左部符号加入集合中
	for _, production := range g.Productions {
		nonTerminals[production.Left.Value] = true
	}

	// 将集合中的符号转换为切片并返回
	result := make([]Symbol, 0, len(nonTerminals))
	for nt := range nonTerminals {
		result = append(result, Symbol{Value: nt, IsTerminal: false})
	}
	return result
}

// GetTerminals
//获取输出的文法所有产生式的终结符
func (g *Grammar) GetTerminals() []Symbol {
	terminals := make(map[string]bool)
	// 遍历产生式右部得到所有符号
	for _, production := range g.Productions {
		for _, alternative := range production.Right {
			for _, sym := range alternative.Symbols {
				// 如果符号不是非终结符，则将其加入终结符集合
				terminals[sym.Value] = true
			}
		}
	}

	// 从终结符集合中删除非终结符
	nonTerminals := g.GetNonTerminals()
	for _, nt := range nonTerminals {
		delete(terminals, nt.Value)
	}
	// 将集合中的终结符转换为切片并返回
	result := make([]Symbol, 0, len(terminals))
	for t := range terminals {
		result = append(result, Symbol{Value: t, IsTerminal: true})
	}
	return result
}

// AllNullable
//得到字符串的可空性
func (g *Grammar) AllNullable(symbols []Symbol) bool {
	for _, symbol := range symbols {
		if symbol.IsTerminal || !g.Nullable[symbol.Value] {
			return false
		}
	}
	return true
}

// GetFirst
//得到字符串的first集
func (g Grammar) GetFirst(symbols []Symbol) []Symbol {
	result := []Symbol{}
	for _, symbol := range symbols {
		if symbol.IsTerminal {
			result = append(result, symbol)
			break
		} else {
			first, exists := g.FirstSet[symbol]
			if exists {
				for s := range first {
					if s.Value != "ε" {
						result = append(result, s)
					}
				}
			}
		}
	}
	if g.AllNullable(symbols) {
		result = append(result, Symbol{Value: "ε", IsTerminal: true})
	}
	return result
}

// Select
//select集是对每个产生式进行处理
//1.select(S->ab)=first(a)
//select(S->AB)，若AB能得出->ε，则select(S->AB)={first(AB)-{ε}}∪follow(S)。反之，select(S->AB)=first(AB)
func (g Grammar) Select(left Symbol, right []Symbol) []Symbol {
	result := []Symbol{}
	if right[0].IsTerminal {
		result = append(result, right[0])
		return result
	}
	if g.AllNullable(right) {
		for s := range g.FollowSet[left] {
			result = append(result, s)
		}
		for _, s := range g.GetFirst(right) {
			if s.Value != "v" {
				result = append(result, s)
			}
		}

	} else {
		for _, s := range g.GetFirst(right) {
			result = append(result, s)
		}
	}
	return result
}

//分析栈和字符栈，懒得写个栈结构了，用切片将就吧，characterStack默认切片首元素为栈顶，尾元素为栈底。analysisStack默认切片首元素为栈底，尾元素为栈顶
//在循环中，检查分析栈顶的符号。如果它是一个终结符，请检查它是否与 characterStack 的栈顶元素匹配。如果匹配，则将两个栈的栈顶元素弹出；如果不匹配，则输出错误消息并返回。如果匹配且都为终止符#，则匹配成功
//如果栈顶符号是一个非终结符，请在预测分析表（g.Predict）中查找与当前非终结符和 characterStack 栈顶元素对应的产生式。将产生式右侧的符号逆序压入 analysisStack
func (g Grammar) parse(strs string) {
	fmt.Printf("%s的分析过程\n", strs)
	//计数器，分析的步骤
	count := 1
	//分析栈和字符栈的初始化
	var characterStack []string
	for _, s := range strs {
		characterStack = append(characterStack, string(s))
	}
	characterStack = append(characterStack, "#")
	var analysisStack []Symbol
	analysisStack = append(analysisStack, Symbol{"#", true})
	analysisStack = append(analysisStack, g.Start)
	// 使用 tabwriter 对输出进行对齐
	w := tabwriter.NewWriter(os.Stdout, 8, 0, 2, ' ', 0)
	parseContinue := true
	for parseContinue {
		topAnalysis := analysisStack[len(analysisStack)-1]
		topCharacter := characterStack[0]
		printStep(w, count, analysisStack, characterStack)
		w.Flush()
		fmt.Printf("\t")
		if topAnalysis.IsTerminal {
			if topAnalysis.Value == "#" {
				fmt.Println("输入的字符串分析成功.")
				parseContinue = false
			}
			if topAnalysis.Value == topCharacter {
				fmt.Printf("匹配成功%s.\n", topAnalysis.Value)
				analysisStack = analysisStack[:len(analysisStack)-1]
				characterStack = characterStack[1:]
			} else {
				fmt.Printf("匹配失败.\n")
				parseContinue = false
			}
		} else {
			prod, exist := g.Predict[topAnalysis][Symbol{topCharacter, true}]
			if exist {
				analysisStack = analysisStack[:len(analysisStack)-1]
				newSymbols := []Symbol{}
				for _, s := range prod.Right[0].Symbols {
					newSymbols = append(newSymbols, s)
				}
				for len(newSymbols) > 0 {
					if newSymbols[len(newSymbols)-1].Value != "ε" {
						analysisStack = append(analysisStack, newSymbols[len(newSymbols)-1])
					}
					newSymbols = newSymbols[:len(newSymbols)-1]
				}
				//打印使用的产生式
				fmt.Printf("使用产生式 %s -> ", topAnalysis.Value)
				for _, sym := range prod.Right[0].Symbols {
					fmt.Printf("%s ", sym.Value)
				}
				fmt.Println()
			} else {
				fmt.Printf("匹配失败.\n")
				parseContinue = false
			}
		}
		count++
	}
}
func printStep(w *tabwriter.Writer, step int, analysisStack []Symbol, characterStack []string) {
	fmt.Fprintf(w, "%d\t", step)
	for _, a := range analysisStack {
		fmt.Fprintf(w, "%s", a.Value)
	}
	fmt.Fprint(w, "\t")
	for _, c := range characterStack {
		fmt.Fprintf(w, "%s", c)
	}
}
func main() {
	prods := make([]Production, 0)
	reader := bufio.NewReader(os.Stdin)

	// 输入开始符
	startSymbolStruct := Symbol{}
	for {
		fmt.Print("Enter start symbol: ")
		startSymbol, _ := reader.ReadString('\n')
		startSymbol = strings.TrimSpace(startSymbol)
		if len(startSymbol) == 1 {
			startSymbolStruct = Symbol{Value: startSymbol, IsTerminal: false}
			break
		} else {
			fmt.Println("Start symbol must be a single character.")
		}
	}
	// 输入产生式
	for {
		fmt.Print("Enter production (or q to quit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "q" {
			break
		}
		parts := strings.Split(input, "->")
		if len(parts) != 2 {
			fmt.Println("Invalid production")
			continue
		}
		left := strings.TrimSpace(parts[0])
		rightPorts := strings.Split(strings.TrimSpace(parts[1]), "|")

		alternatives := make([]Alternative, len(rightPorts))
		for i, rightStr := range rightPorts {
			trimmedRightStr := strings.TrimSpace(rightStr)
			symbols := []Symbol{}
			for _, symbolRune := range trimmedRightStr {
				// 假设文法输入时将所有符号视为非终结符
				symbolStr := string(symbolRune)
				symbols = append(symbols, Symbol{Value: symbolStr, IsTerminal: false})
			}
			alternatives[i] = Alternative{Symbols: symbols}
		}

		leftSymbol := Symbol{Value: left, IsTerminal: false}
		prod := Production{Left: leftSymbol, Right: alternatives}
		prods = append(prods, prod)
	}

	// 创建文法 g
	g := Grammar{
		Start:       startSymbolStruct,
		Productions: prods,
		Nullable:    nil,
		FirstSet:    nil,
		FollowSet:   nil,
		Predict:     nil,
	}
	fmt.Println()
	g.PrintGrammar()
	fmt.Println()
	if g.GInit() {
		for {
			fmt.Print("Please enter the string you want to parse (or q to quit): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "q" {
				break
			}
			input = strings.TrimSpace(input)
			fmt.Println()
			g.parse(input)
			fmt.Println()
		}
	}
}

//初始化专区

// GInit
//对于一个输入了开始符和产生式集的文法进行初始化，得到他的Nullable，FirstSet，FollowSet，Predict，并判断是否为LL1文法，返回结果
func (g *Grammar) GInit() bool {
	//更新符号的 IsTerminal 字段
	terminals := g.GetTerminals()
	terminalSet := make(map[string]bool)
	for _, t := range terminals {
		terminalSet[t.Value] = true
	}
	for i, production := range g.Productions {
		for j, alternative := range production.Right {
			for k, sym := range alternative.Symbols {
				if terminalSet[sym.Value] {
					g.Productions[i].Right[j].Symbols[k].IsTerminal = true
				}
			}
		}
	}
	//提取左公因子与消除左递归
	g.extractCommonFactors()
	g.eliminateDirectLeftRecursion()
	g.PrintNonTerminals()
	g.PrintTerminals()
	fmt.Println()

	//初始化nullable表，first表，follow表
	g.initializeNullable()
	g.PrintNullableTable()
	fmt.Println()

	g.initializeFirstSet()
	g.PrintFirstSet()
	fmt.Println()

	g.initializeFollowSet()
	g.PrintFollowSet()
	fmt.Println()
	//如果是LL1文法则继续否则结束
	if g.isLL1() {
		fmt.Println()
		g.initializePredict()
		g.PrintPredict()
		return true
	}
	return false
}

//A→Aα1|Aα2|…|Aαm|β1|β2|…|βn
//消除后为
//A→(β1|β2|…|βn)A’
//A’→(α1|α2|…|αm)A’|ε
func (g *Grammar) eliminateDirectLeftRecursion() {
	isReForm := false
	for _, nt1 := range g.GetNonTerminals() {
		// 遍历每个非终结符nt1
		for j, prod := range g.Productions {
			// 遍历产生式找到该非终结符
			if prod.Left == nt1 {
				alpha := make([]Alternative, 0)
				beta := make([]Alternative, 0)

				// 查找是否有左递归
				for _, alt := range prod.Right {
					if len(alt.Symbols) > 0 && alt.Symbols[0].Value == nt1.Value {
						alpha = append(alpha, Alternative{Symbols: alt.Symbols[1:]})
					} else {
						beta = append(beta, alt)
					}
				}

				// 有左递归，需要消除
				if len(alpha) > 0 {
					isReForm = true
					newNt1 := Symbol{Value: nt1.Value + "'", IsTerminal: false}

					// 更新原有产生式
					updatedRight := make([]Alternative, 0)
					for _, b := range beta {
						updatedRight = append(updatedRight, Alternative{Symbols: append(b.Symbols, newNt1)})
					}
					g.Productions[j].Right = updatedRight

					// 添加新产生式
					newRight := make([]Alternative, 0)
					for _, a := range alpha {
						newRight = append(newRight, Alternative{Symbols: append(a.Symbols, newNt1)})
					}
					newRight = append(newRight, Alternative{Symbols: []Symbol{{Value: "ε", IsTerminal: true}}})
					g.Productions = append(g.Productions, Production{Left: newNt1, Right: newRight})
				}
			}
		}
	}

	if isReForm {
		fmt.Println("eliminateDirectLeftRecursion grammar:")
		for _, prod := range g.Productions {
			rightParts := make([]string, len(prod.Right))
			for i, alt := range prod.Right {
				symbols := make([]string, len(alt.Symbols))
				for j, sym := range alt.Symbols {
					symbols[j] = sym.Value
				}
				rightParts[i] = strings.Join(symbols, "")
			}
			fmt.Printf("%s -> %s\n", prod.Left.Value, strings.Join(rightParts, "|"))
		}
		fmt.Println()
	}
}

//提取公因子：将产生式中的公共前缀提取出来，简化文法。
func (g *Grammar) extractCommonFactors() {
	flag := false
	// 计数器，用于生成新的非终结符
	newNonTerminalCounter := 0
	for i, production := range g.Productions {
		// 新的备选项列表，用于存储提取公共前缀后的备选项
		newAlternatives := []Alternative{}
		// 创建一个映射，用于将具有相同前缀的备选项归类到一起。
		// 键是前缀（即符号的值），值是具有相同前缀的备选项列表。
		prefixMap := make(map[Symbol][]Alternative)
		// 遍历产生式的每个备选项,，存入前缀映射
		for _, alternative := range production.Right {
			if len(alternative.Symbols) > 0 {
				// 取备选项的第一个符号作为键
				firstSymbol := alternative.Symbols[0]
				// 将具有相同前缀的备选项添加到映射中
				if _, ok := prefixMap[firstSymbol]; !ok {
					prefixMap[firstSymbol] = []Alternative{}
				}
				prefixMap[firstSymbol] = append(prefixMap[firstSymbol], alternative)
			}
		}
		// 遍历具有相同前缀的备选项组
		for _, alternatives := range prefixMap {
			// 如果具有相同前缀的备选项数量大于 1
			if len(alternatives) > 1 {
				flag = true
				// 查找多个备选项的最长公共前缀
				commonPrefix := findLongestCommonPrefix(alternatives)
				// 生成一个新的非终结符，例如 "A1"、"A2" 等
				newNonTerminalCounter++
				newValue := fmt.Sprintf("A%d", newNonTerminalCounter)
				newNonTerminal := Symbol{
					Value:      newValue,
					IsTerminal: false,
				}
				// 移除原来备选项中的公共前缀
				newRight := removeCommonPrefix(alternatives, commonPrefix)
				// 创建一个新的产生式，左侧是新的非终结符，右侧是移除公共前缀后的备选项列表
				newProduction := Production{
					Left:  newNonTerminal,
					Right: newRight,
				}
				// 将新的产生式添加到文法的产生式列表中
				g.Productions = append(g.Productions, newProduction)
				// 创建一个新的备选项，包括公共前缀和新的非终结符
				var newAlternative Alternative
				commonPrefixCopy := make([]Symbol, len(commonPrefix))
				copy(commonPrefixCopy, commonPrefix)
				newAlternative.Symbols = commonPrefixCopy
				newAlternative.Symbols = append(newAlternative.Symbols, newNonTerminal)
				// 将新的备选项添加到新的备选项列表中
				newAlternatives = append(newAlternatives, newAlternative)
			} else {
				// 如果只有一个备选项，直接将其添加到新的备选项列表中
				newAlternatives = append(newAlternatives, alternatives[0])
			}
		}
		// 更新产生式的备选项列表
		g.Productions[i].Right = newAlternatives
	}
	if flag {
		fmt.Println("extractCommonFactors grammar:")
		for _, prod := range g.Productions {
			rightParts := make([]string, len(prod.Right))
			for i, alt := range prod.Right {
				symbols := make([]string, len(alt.Symbols))
				for j, sym := range alt.Symbols {
					symbols[j] = sym.Value
				}
				rightParts[i] = strings.Join(symbols, "")
			}
			fmt.Printf("%s -> %s\n", prod.Left.Value, strings.Join(rightParts, "|"))
		}
		fmt.Println()
	}
}
func findLongestCommonPrefix(alternatives []Alternative) []Symbol {
	if len(alternatives) == 0 {
		return nil
	}
	// 获取第一个备选项的符号列表作为参考
	reference := alternatives[0].Symbols
	// 初始化最长公共前缀的长度为参考备选项的长度
	maxLength := len(reference)
	// 遍历其余的备选项
	for _, alternative := range alternatives[1:] {
		i := 0
		// 比较当前备选项与参考备选项的符号，找出公共前缀的长度存为i
		for i < len(alternative.Symbols) && i < maxLength {
			if alternative.Symbols[i] != reference[i] {
				break
			}
			i++
		}

		// 更新最长公共前缀的长度
		if i < maxLength {
			maxLength = i
		}
	}

	// 返回最长公共前缀的符号列表
	return reference[:maxLength]
}
func removeCommonPrefix(alternatives []Alternative, prefix []Symbol) []Alternative {
	// 创建一个新的备选项列表，用于存储移除公共前缀后的备选项
	newAlternatives := []Alternative{}
	// 获取公共前缀的长度
	prefixLength := len(prefix)
	// 遍历原始备选项列表
	for _, alternative := range alternatives {
		// 创建一个新的符号列表，用于存储移除公共前缀后的符号
		newSymbols := []Symbol{}

		// 如果原始符号列表的长度大于公共前缀的长度，则从原始符号列表中移除公共前缀
		if len(alternative.Symbols) > prefixLength {
			newSymbols = alternative.Symbols[prefixLength:]
		}

		// 将移除公共前缀后的符号列表添加到新的备选项中
		newAlternative := Alternative{Symbols: newSymbols}

		// 将新的备选项添加到新的备选项列表中
		newAlternatives = append(newAlternatives, newAlternative)
	}

	// 返回移除公共前缀后的新的备选项列表
	return newAlternatives
}

//以S->ε为基础循环遍历产生式，找到所有能推导出空串的非终结符存入空串表
func (g *Grammar) initializeNullable() {
	g.Nullable = make(map[string]bool)
	// 初始化，将所有非终结符设置为不可空
	for _, production := range g.Productions {
		// 设置非终结符的可空性为 false
		g.Nullable[production.Left.Value] = false
	}

	// 反复遍历产生式，直到没有新的可空非终结符被添加
	// 使用一个布尔值变量 `changed` 来跟踪是否有新的可空非终结符被找到
	changed := true
	for changed {
		// 设置 `changed` 为 false，表示在本轮遍历中尚未找到新的可空非终结符
		changed = false

		// 遍历产生式
		for _, production := range g.Productions {
			// 如果当前非终结符尚未被标记为可空
			if !g.Nullable[production.Left.Value] {
				// 遍历产生式右侧的替代项
				for _, alternative := range production.Right {
					// 假定当前替代项可导出空串
					nullable := true

					// 遍历替代项中的符号
					for _, symbol := range alternative.Symbols {
						// 如果符号是终结符，或者符号是一个不可空的非终结符
						if symbol.Value == "ε" {
							//空串，直接跳出
							break
						}
						if symbol.IsTerminal || !g.Nullable[symbol.Value] {
							// 将 `nullable` 设置为 false，表示当前替代项不可导出空串
							nullable = false
							break
						}
					}
					// 如果当前替代项可导出空串
					if nullable {
						// 将左侧非终结符标记为可空
						g.Nullable[production.Left.Value] = true
						// 将 `changed` 设置为 true，表示在本轮遍历中找到了新的可空非终结符
						changed = true
						// 跳出替代项遍历循环
						break
					}
				}
			}
		}
	}
}

//终结符的first集为自己
//非终结符的first集并入（除ε之外）
//如果可空继续判断下一个字符，如果都可空则加入ε
func (g *Grammar) initializeFirstSet() {
	g.FirstSet = make(map[Symbol]map[Symbol]bool)

	// 遍历产生式，初始化每个符号的 First 集
	for _, production := range g.Productions {
		if _, ok := g.FirstSet[production.Left]; !ok {
			g.FirstSet[production.Left] = make(map[Symbol]bool)
		}
		for _, alternative := range production.Right {
			for _, symbol := range alternative.Symbols {
				if _, ok := g.FirstSet[symbol]; !ok {
					g.FirstSet[symbol] = make(map[Symbol]bool)
				}
			}
		}
	}

	// 将终结符添加到它们自己的 First 集中
	for symbol, symbolFirstSet := range g.FirstSet {
		if symbol.Value != "ε" && symbol.IsTerminal {
			symbolFirstSet[symbol] = true
		}
	}

	// 计算非终结符的 First 集
	changed := true // 反复遍历产生式，直到没有新的符号被添加到 First 集
	for changed {
		changed = false
		for _, production := range g.Productions {
			left := production.Left
			for _, alternative := range production.Right {
				nullable := true
				for _, symbol := range alternative.Symbols {
					// 如果符号是非终结符
					if !symbol.IsTerminal {
						// 将 symbol 的 First 集合并到 left 的 First 集中
						for s, exist := range g.FirstSet[symbol] {
							if exist && s.Value != "ε" && !g.FirstSet[left][s] {
								g.FirstSet[left][s] = true
								changed = true
							}
						}
						// 如果当前符号是不可空的，跳出循环
						if !g.Nullable[symbol.Value] {
							nullable = false
							break
						}
					} else {
						// 如果符号是终结符，将其添加到 left 的 First 集中，并跳出循环
						if !g.FirstSet[left][symbol] {
							g.FirstSet[left][symbol] = true
							changed = true
						}
						nullable = false
						break
					}
				}
				// 如果所有符号都是可空的，则将 "ε" 添加到 left 的 First 集中
				if nullable {
					if !g.FirstSet[left][Symbol{Value: "ε", IsTerminal: false}] {
						g.FirstSet[left][Symbol{Value: "ε", IsTerminal: false}] = true
						changed = true
					}
				}
			}
		}
	}
}

//开始符号的follow应该有输入结束语#
//终结符没有follow集
//对于非终结符 A，如果 A 后面紧跟着一个终结符 a，则将 a 添加到 A 的 Follow 集中。
//对于非终结符 A，如果 A 后面紧跟着一个非终结符 B，则将 B 的 First 集（不包括 "ε"）中的所有符号添加到 A 的 Follow 集中。
//对于非终结符 A，如果 A 后面紧跟着一个非终结符 B，且 B 可导出空串（"ε"），则将产生式左侧非终结符的 Follow 集中的所有符号添加到 A 的 Follow 集中。
func (g *Grammar) initializeFollowSet() {
	g.FollowSet = make(map[Symbol]map[Symbol]bool)
	// 初始化非终结符的 Follow 集
	for _, production := range g.Productions {
		left := production.Left
		if _, ok := g.FollowSet[left]; !ok {
			g.FollowSet[left] = make(map[Symbol]bool)
		}
	}
	// 将文法开始符号的 Follow 集设为 { # }，表示输入结束符号
	g.FollowSet[g.Start] = map[Symbol]bool{Symbol{Value: "#", IsTerminal: true}: true}
	// 反复遍历产生式，直到 Follow 集不再发生变化
	changed := true
	for changed {
		changed = false
		for _, production := range g.Productions {
			left := production.Left
			//遍历产生式，得到symbol为非终结符A
			for _, alternative := range production.Right {
				for i, symbol := range alternative.Symbols {
					if !symbol.IsTerminal {
						for j := i + 1; j < len(alternative.Symbols); j++ {
							nextSymbol := alternative.Symbols[j]
							if nextSymbol.IsTerminal {
								//终结符，加入
								if _, exists := g.FollowSet[symbol][nextSymbol]; !exists {
									g.FollowSet[symbol][nextSymbol] = true
									changed = true
								}
								break
							} else {
								//将 nextSymbol 的 First 集合加入 symbol 的 Follow 集合
								for firstSymbol := range g.FirstSet[nextSymbol] {
									if firstSymbol.Value != "ε" {
										if _, exists := g.FollowSet[symbol][firstSymbol]; !exists {
											g.FollowSet[symbol][firstSymbol] = true
											changed = true
										}
									}
								}
								// 如果 nextSymbol 可以为空，则继续往后处理
								if !g.Nullable[nextSymbol.Value] {
									break
								}
							}
						}
						// 如果非终结符 A 后面的所有非终结符都可以为空，则将产生式左侧非终结符的 Follow 集中的所有符号添加到 A 的 Follow 集中
						if g.AllNullable(alternative.Symbols[i+1:]) {
							for followSymbol := range g.FollowSet[left] {
								if _, exists := g.FollowSet[symbol][followSymbol]; !exists {
									g.FollowSet[symbol][followSymbol] = true
									changed = true
								}
							}
						}
					}
				}
			}
		}
	}
}

//Predict     map[Symbol]map[Symbol]Production
//构造分析表的第一个元素为左边的非终结符，第二个元素为上面的终结符
//对文法G的每个产生式A->α 执行如下步骤：
//（1）对每个a∈First(α)，把 A->α 加入M[A,a]
//（2）若 ε∈First(α)，则对任何b∈Follow(A) ,把 A-> ε加至M[A,b]中
//得到构造表Predict 存储了M[A,b]
func (g *Grammar) initializePredict() {
	var epsilon Symbol = Symbol{
		Value:      "ε",
		IsTerminal: true,
	}
	g.Predict = make(map[Symbol]map[Symbol]Production)
	//初始化
	for _, nonterminal := range g.GetNonTerminals() {
		g.Predict[nonterminal] = make(map[Symbol]Production)
	}
	for _, prod := range g.Productions {
		for _, alter := range prod.Right {
			//对于每个产生式prod.left->alter
			for _, s := range g.GetFirst(alter.Symbols) {
				g.Predict[prod.Left][s] = Production{
					Left:  prod.Left,
					Right: []Alternative{alter},
				}
			}
			if findString(g.GetFirst(alter.Symbols), "ε") {
				for s := range g.FollowSet[prod.Left] {
					g.Predict[prod.Left][s] = Production{
						Left: prod.Left,
						Right: []Alternative{
							{[]Symbol{epsilon}},
						},
					}
				}
			}
		}
	}
}
func findString(strs []Symbol, s string) bool {
	for _, str := range strs {
		if str.Value == s {
			return true
		}
	}
	return false
}

//isLL1
//对于每个非终结符A，检查每对产生式P1, P2是否存在以下情况
//1. P1和P2都以相同的终结符开始，这是FIRST集冲突
//2. P1的FIRST集包含空串，P2的FIRST集与A的FOLLOW集有交集，这是FOLLOW集冲突
func (g Grammar) isLL1() bool {
	fmt.Println(" LL1 grammar or not:")
	flag := false
	for _, prod := range g.Productions {
		if len(prod.Right) > 1 {
			for i := 0; i < len(prod.Right); i++ {
				for j := i + 1; j < len(prod.Right); j++ {
					select1 := g.Select(prod.Left, prod.Right[i].Symbols)
					select2 := g.Select(prod.Left, prod.Right[j].Symbols)
					if hasIntersection(select1, select2) {
						flag = true
						fmt.Printf("select(%s)∩select(%s)!=Ø\n", symbolsToString(prod.Right[i].Symbols), symbolsToString(prod.Right[j].Symbols))
					} else {
						fmt.Printf("select(%s)∩select(%s)=Ø\n", symbolsToString(prod.Right[i].Symbols), symbolsToString(prod.Right[j].Symbols))
					}
				}
			}
		}
	}
	if flag {
		fmt.Println("The grammar you entered is not the LL1 grammar")
	} else {
		fmt.Println("The grammar you entered is  the LL1 grammar,please continue")
	}
	return !flag
}
func hasIntersection(slice1, slice2 []Symbol) bool {
	elementMap := make(map[Symbol]bool)
	for _, elem := range slice1 {
		elementMap[elem] = true
	}
	// 遍历slice2，查找是否存在于elementMap中的元素
	for _, elem := range slice2 {
		if _, found := elementMap[elem]; found {
			return true
		}
	}
	return false
}
func symbolsToString(symbols []Symbol) string {
	values := make([]string, len(symbols))
	for i, symbol := range symbols {
		values[i] = symbol.Value
	}
	return strings.Join(values, "")
}

//打印专区

func (g *Grammar) PrintGrammar() {
	fmt.Println("Input grammar:")
	for _, prod := range g.Productions {
		rightParts := make([]string, len(prod.Right))
		for i, alt := range prod.Right {
			symbols := make([]string, len(alt.Symbols))
			for j, sym := range alt.Symbols {
				symbols[j] = sym.Value
			}
			rightParts[i] = strings.Join(symbols, "")
		}
		fmt.Printf("%s -> %s\n", prod.Left.Value, strings.Join(rightParts, "|"))
	}
}
func (g *Grammar) PrintNonTerminals() {
	nonTerminals := g.GetNonTerminals()
	fmt.Println("Nonterminals:")
	printSymbolSlice(nonTerminals)
}
func (g *Grammar) PrintTerminals() {
	terminals := g.GetTerminals()
	fmt.Println("Terminals:")
	printSymbolSlice(terminals)
}
func printSymbolSlice(symbols []Symbol) {
	fmt.Printf("[ ")
	for _, sym := range symbols {
		fmt.Printf("%s ", sym.Value)
	}
	fmt.Printf("]\n")
}
func (g *Grammar) PrintNullableTable() {
	fmt.Println("Nullable Table:")

	for nonTerminal, nullable := range g.Nullable {
		fmt.Printf("%s: %v\n", nonTerminal, nullable)
	}
}
func (g *Grammar) PrintFirstSet() {
	fmt.Println("First Sets:")

	for symbol, symbolFirstSet := range g.FirstSet {
		// 跳过 "ε" 符号
		if symbol.IsTerminal {
			continue
		}
		fmt.Printf("First(%s) = {", symbol.Value)
		first := true
		for s, present := range symbolFirstSet {
			if present {
				if !first {
					//如果不是第一个输出的符号，那么在输出符号前添加逗号和空格。
					fmt.Print(", ")
				}
				fmt.Print(s.Value)
				first = false
			}
		}
		fmt.Println("}")
	}
}
func (g *Grammar) PrintFollowSet() {
	fmt.Println("Follow Sets:")

	for symbol, symbolFollowSet := range g.FollowSet {
		fmt.Printf("Follow(%s) = {", symbol.Value)
		first := true
		for s, present := range symbolFollowSet {
			if present {
				if !first {
					//如果不是第一个输出的符号，那么在输出符号前添加逗号和空格。
					fmt.Print(", ")
				}
				fmt.Print(s.Value)
				first = false
			}
		}
		fmt.Println("}")
	}
}
func (g *Grammar) PrintPredict() {
	fmt.Println("Predict Table:")
	// 获取所有非终结符
	nonTerminals := g.GetNonTerminals()
	// 获取所有终结符
	terminals := g.GetTerminals()
	terminals = deleteSymbol(terminals, "ε")
	terminals = addSymbol(terminals, "#", true)
	// 使用 tabwriter 对输出进行对齐
	w := tabwriter.NewWriter(os.Stdout, 8, 0, 2, ' ', 0)
	// 打印表头
	fmt.Fprint(w, "\t")
	for _, t := range terminals {
		fmt.Fprintf(w, "%s\t", t.Value)
	}
	fmt.Fprintln(w)
	// 遍历非终结符
	for _, nonTerminal := range nonTerminals {
		fmt.Fprintf(w, "%s\t", nonTerminal.Value)
		// 遍历终结符
		for _, terminal := range terminals {
			production, exists := g.Predict[nonTerminal][terminal]
			if exists {
				// 打印产生式
				fmt.Fprintf(w, "%s -> ", production.Left.Value)
				for _, symbol := range production.Right[0].Symbols {
					fmt.Fprint(w, symbol.Value)
				}
			} else {
				fmt.Fprint(w, "    ") // 空产生式用空格填充
			}
			fmt.Fprint(w, "\t")
		}
		fmt.Fprintln(w)
	}
	// 刷新 tabwriter 输出
	w.Flush()
}
func deleteSymbol(symbols []Symbol, value string) []Symbol {
	for i := range symbols {
		if symbols[i].Value == value {
			symbols = append(symbols[:i], symbols[i+1:]...)
			break
		}
	}
	return symbols
}
func addSymbol(symbols []Symbol, value string, isTerminal bool) []Symbol {
	for i := range symbols {
		if symbols[i].Value == value {
			return symbols
		}
	}
	symbols = append(symbols, Symbol{Value: value, IsTerminal: isTerminal})
	return symbols
}
