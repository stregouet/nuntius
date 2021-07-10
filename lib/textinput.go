package lib

func InsertRune(input []rune, cursorpos int, newchr rune) []rune {
	if cursorpos == len(input) {
		return append(input, newchr)
	}
	result := make([]rune, len(input)+1)
	copy(result, input[:cursorpos])
	result[cursorpos] = newchr
	copy(result[cursorpos+1:], input[cursorpos:])
	return result
}

func RemoveRuneBackward(input []rune, cursorpos int) []rune {
	if cursorpos == len(input) {
		return input[:cursorpos-1]
	}
	return append(input[:cursorpos-1], input[cursorpos:]...)
}

func RemoveRuneWordBackward(input []rune, cursorpos int) ([]rune, int) {
	words := make([][]rune, 0)
	curword := make([]rune, 0)
	for pos, chr := range input {
		if pos == cursorpos {
			// stop loop without adding last word
			break
		}
		if len(curword) > 1 {
			// if last chr in curword is a delimiter we should
			// append word and create an empty []rune
			switch curword[len(curword)-1] {
			case ' ', '\'', ':':
				words = append(words, curword)
				curword = make([]rune, 0)
			}
		}
		curword = append(curword, chr)
	}
	res := make([]rune, 0)
	for _, w := range words {
		res = append(res, w...)
	}
	return append(res, input[cursorpos:]...), len(curword)
}
