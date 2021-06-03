package lib

import (
	"errors"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var UnknownKeyStroke = errors.New("unknown keystroke")
var keyNamesToStroke map[string]KeyStroke
var keyNames map[tcell.Key]string

type KeyStroke struct {
	Key  tcell.Key
	Rune rune
}

func (k *KeyStroke) String() string {
	if k.Key == tcell.KeyRune {
		if unicode.IsUpper(k.Rune) {
			return "s-" + string(unicode.ToLower(k.Rune))
		}
		return string(k.Rune)
	}
	return keyNames[k.Key]
}

func KeyStrokesToString(ks []*KeyStroke) string {
	parts := make([]string, 0)
	for _, k := range ks {
		parts = append(parts, k.String())
	}
	return strings.Join(parts, " ")
}

func KeyStrokeFromEvent(ev tcell.Event) *KeyStroke {
	if evk, ok := ev.(*tcell.EventKey); !ok {
		return nil
	} else {
		if evk.Key() == tcell.KeyRune {
			return &KeyStroke{tcell.KeyRune, evk.Rune()}
		} else {
			return &KeyStroke{evk.Key(), 0}
		}
	}
}


func ParseKeyStroke(userinput string) ([]*KeyStroke, error) {
	result := make([]*KeyStroke, 0)
	for _, part := range strings.Split(strings.Trim(userinput, " "), " ") {
		if k, ok := keyNamesToStroke[part]; ok {
			result = append(result, &k)
		} else if runewidth.StringWidth(part) == 1 {
			r := []rune(part)
			result = append(result, &KeyStroke{tcell.KeyRune, r[0]})
		} else if runewidth.StringWidth(part) == 3 && part[:2] == "s-" {
			r := []rune(part)
			result = append(result, &KeyStroke{tcell.KeyRune, unicode.ToUpper(r[2])})
		} else {
			return nil, UnknownKeyStroke
		}
	}
	return result, nil
}

func init() {
	keyNamesToStroke = make(map[string]KeyStroke)
	keyNamesToStroke["space"] = KeyStroke{tcell.KeyRune, ' '}
	keyNamesToStroke["enter"] = KeyStroke{tcell.KeyEnter, 0}
	keyNamesToStroke["up"] = KeyStroke{tcell.KeyUp, 0}
	keyNamesToStroke["down"] = KeyStroke{tcell.KeyDown, 0}
	keyNamesToStroke["right"] = KeyStroke{tcell.KeyRight, 0}
	keyNamesToStroke["left"] = KeyStroke{tcell.KeyLeft, 0}
	keyNamesToStroke["upleft"] = KeyStroke{tcell.KeyUpLeft, 0}
	keyNamesToStroke["upright"] = KeyStroke{tcell.KeyUpRight, 0}
	keyNamesToStroke["downleft"] = KeyStroke{tcell.KeyDownLeft, 0}
	keyNamesToStroke["downright"] = KeyStroke{tcell.KeyDownRight, 0}
	keyNamesToStroke["center"] = KeyStroke{tcell.KeyCenter, 0}
	keyNamesToStroke["pgup"] = KeyStroke{tcell.KeyPgUp, 0}
	keyNamesToStroke["pgdn"] = KeyStroke{tcell.KeyPgDn, 0}
	keyNamesToStroke["home"] = KeyStroke{tcell.KeyHome, 0}
	keyNamesToStroke["end"] = KeyStroke{tcell.KeyEnd, 0}
	keyNamesToStroke["insert"] = KeyStroke{tcell.KeyInsert, 0}
	keyNamesToStroke["delete"] = KeyStroke{tcell.KeyDelete, 0}
	keyNamesToStroke["help"] = KeyStroke{tcell.KeyHelp, 0}
	keyNamesToStroke["exit"] = KeyStroke{tcell.KeyExit, 0}
	keyNamesToStroke["clear"] = KeyStroke{tcell.KeyClear, 0}
	keyNamesToStroke["cancel"] = KeyStroke{tcell.KeyCancel, 0}
	keyNamesToStroke["print"] = KeyStroke{tcell.KeyPrint, 0}
	keyNamesToStroke["pause"] = KeyStroke{tcell.KeyPause, 0}
	keyNamesToStroke["backtab"] = KeyStroke{tcell.KeyBacktab, 0}
	keyNamesToStroke["f1"] = KeyStroke{tcell.KeyF1, 0}
	keyNamesToStroke["f2"] = KeyStroke{tcell.KeyF2, 0}
	keyNamesToStroke["f3"] = KeyStroke{tcell.KeyF3, 0}
	keyNamesToStroke["f4"] = KeyStroke{tcell.KeyF4, 0}
	keyNamesToStroke["f5"] = KeyStroke{tcell.KeyF5, 0}
	keyNamesToStroke["f6"] = KeyStroke{tcell.KeyF6, 0}
	keyNamesToStroke["f7"] = KeyStroke{tcell.KeyF7, 0}
	keyNamesToStroke["f8"] = KeyStroke{tcell.KeyF8, 0}
	keyNamesToStroke["f9"] = KeyStroke{tcell.KeyF9, 0}
	keyNamesToStroke["f10"] = KeyStroke{tcell.KeyF10, 0}
	keyNamesToStroke["f11"] = KeyStroke{tcell.KeyF11, 0}
	keyNamesToStroke["f12"] = KeyStroke{tcell.KeyF12, 0}
	keyNamesToStroke["f13"] = KeyStroke{tcell.KeyF13, 0}
	keyNamesToStroke["f14"] = KeyStroke{tcell.KeyF14, 0}
	keyNamesToStroke["f15"] = KeyStroke{tcell.KeyF15, 0}
	keyNamesToStroke["f16"] = KeyStroke{tcell.KeyF16, 0}
	keyNamesToStroke["f17"] = KeyStroke{tcell.KeyF17, 0}
	keyNamesToStroke["f18"] = KeyStroke{tcell.KeyF18, 0}
	keyNamesToStroke["f19"] = KeyStroke{tcell.KeyF19, 0}
	keyNamesToStroke["f20"] = KeyStroke{tcell.KeyF20, 0}
	keyNamesToStroke["f21"] = KeyStroke{tcell.KeyF21, 0}
	keyNamesToStroke["f22"] = KeyStroke{tcell.KeyF22, 0}
	keyNamesToStroke["f23"] = KeyStroke{tcell.KeyF23, 0}
	keyNamesToStroke["f24"] = KeyStroke{tcell.KeyF24, 0}
	keyNamesToStroke["f25"] = KeyStroke{tcell.KeyF25, 0}
	keyNamesToStroke["f26"] = KeyStroke{tcell.KeyF26, 0}
	keyNamesToStroke["f27"] = KeyStroke{tcell.KeyF27, 0}
	keyNamesToStroke["f28"] = KeyStroke{tcell.KeyF28, 0}
	keyNamesToStroke["f29"] = KeyStroke{tcell.KeyF29, 0}
	keyNamesToStroke["f30"] = KeyStroke{tcell.KeyF30, 0}
	keyNamesToStroke["f31"] = KeyStroke{tcell.KeyF31, 0}
	keyNamesToStroke["f32"] = KeyStroke{tcell.KeyF32, 0}
	keyNamesToStroke["f33"] = KeyStroke{tcell.KeyF33, 0}
	keyNamesToStroke["f34"] = KeyStroke{tcell.KeyF34, 0}
	keyNamesToStroke["f35"] = KeyStroke{tcell.KeyF35, 0}
	keyNamesToStroke["f36"] = KeyStroke{tcell.KeyF36, 0}
	keyNamesToStroke["f37"] = KeyStroke{tcell.KeyF37, 0}
	keyNamesToStroke["f38"] = KeyStroke{tcell.KeyF38, 0}
	keyNamesToStroke["f39"] = KeyStroke{tcell.KeyF39, 0}
	keyNamesToStroke["f40"] = KeyStroke{tcell.KeyF40, 0}
	keyNamesToStroke["f41"] = KeyStroke{tcell.KeyF41, 0}
	keyNamesToStroke["f42"] = KeyStroke{tcell.KeyF42, 0}
	keyNamesToStroke["f43"] = KeyStroke{tcell.KeyF43, 0}
	keyNamesToStroke["f44"] = KeyStroke{tcell.KeyF44, 0}
	keyNamesToStroke["f45"] = KeyStroke{tcell.KeyF45, 0}
	keyNamesToStroke["f46"] = KeyStroke{tcell.KeyF46, 0}
	keyNamesToStroke["f47"] = KeyStroke{tcell.KeyF47, 0}
	keyNamesToStroke["f48"] = KeyStroke{tcell.KeyF48, 0}
	keyNamesToStroke["f49"] = KeyStroke{tcell.KeyF49, 0}
	keyNamesToStroke["f50"] = KeyStroke{tcell.KeyF50, 0}
	keyNamesToStroke["f51"] = KeyStroke{tcell.KeyF51, 0}
	keyNamesToStroke["f52"] = KeyStroke{tcell.KeyF52, 0}
	keyNamesToStroke["f53"] = KeyStroke{tcell.KeyF53, 0}
	keyNamesToStroke["f54"] = KeyStroke{tcell.KeyF54, 0}
	keyNamesToStroke["f55"] = KeyStroke{tcell.KeyF55, 0}
	keyNamesToStroke["f56"] = KeyStroke{tcell.KeyF56, 0}
	keyNamesToStroke["f57"] = KeyStroke{tcell.KeyF57, 0}
	keyNamesToStroke["f58"] = KeyStroke{tcell.KeyF58, 0}
	keyNamesToStroke["f59"] = KeyStroke{tcell.KeyF59, 0}
	keyNamesToStroke["f60"] = KeyStroke{tcell.KeyF60, 0}
	keyNamesToStroke["f61"] = KeyStroke{tcell.KeyF61, 0}
	keyNamesToStroke["f62"] = KeyStroke{tcell.KeyF62, 0}
	keyNamesToStroke["f63"] = KeyStroke{tcell.KeyF63, 0}
	keyNamesToStroke["f64"] = KeyStroke{tcell.KeyF64, 0}
	keyNamesToStroke["c-space"] = KeyStroke{tcell.KeyCtrlSpace, 0}
	keyNamesToStroke["c-a"] = KeyStroke{tcell.KeyCtrlA, 0}
	keyNamesToStroke["c-b"] = KeyStroke{tcell.KeyCtrlB, 0}
	keyNamesToStroke["c-c"] = KeyStroke{tcell.KeyCtrlC, 0}
	keyNamesToStroke["c-d"] = KeyStroke{tcell.KeyCtrlD, 0}
	keyNamesToStroke["c-e"] = KeyStroke{tcell.KeyCtrlE, 0}
	keyNamesToStroke["c-f"] = KeyStroke{tcell.KeyCtrlF, 0}
	keyNamesToStroke["c-g"] = KeyStroke{tcell.KeyCtrlG, 0}
	keyNamesToStroke["c-h"] = KeyStroke{tcell.KeyCtrlH, 0}
	keyNamesToStroke["c-i"] = KeyStroke{tcell.KeyCtrlI, 0}
	keyNamesToStroke["c-j"] = KeyStroke{tcell.KeyCtrlJ, 0}
	keyNamesToStroke["c-k"] = KeyStroke{tcell.KeyCtrlK, 0}
	keyNamesToStroke["c-l"] = KeyStroke{tcell.KeyCtrlL, 0}
	keyNamesToStroke["c-m"] = KeyStroke{tcell.KeyCtrlM, 0}
	keyNamesToStroke["c-n"] = KeyStroke{tcell.KeyCtrlN, 0}
	keyNamesToStroke["c-o"] = KeyStroke{tcell.KeyCtrlO, 0}
	keyNamesToStroke["c-p"] = KeyStroke{tcell.KeyCtrlP, 0}
	keyNamesToStroke["c-q"] = KeyStroke{tcell.KeyCtrlQ, 0}
	keyNamesToStroke["c-r"] = KeyStroke{tcell.KeyCtrlR, 0}
	keyNamesToStroke["c-s"] = KeyStroke{tcell.KeyCtrlS, 0}
	keyNamesToStroke["c-t"] = KeyStroke{tcell.KeyCtrlT, 0}
	keyNamesToStroke["c-u"] = KeyStroke{tcell.KeyCtrlU, 0}
	keyNamesToStroke["c-v"] = KeyStroke{tcell.KeyCtrlV, 0}
	keyNamesToStroke["c-w"] = KeyStroke{tcell.KeyCtrlW, 0}
	keyNamesToStroke["c-x"] = KeyStroke{tcell.KeyCtrlX, rune(tcell.KeyCAN)}
	keyNamesToStroke["c-y"] = KeyStroke{tcell.KeyCtrlY, 0} // TODO: runes for the rest
	keyNamesToStroke["c-z"] = KeyStroke{tcell.KeyCtrlZ, 0}
	keyNamesToStroke["c-]"] = KeyStroke{tcell.KeyCtrlLeftSq, 0}
	keyNamesToStroke["c-\\"] = KeyStroke{tcell.KeyCtrlBackslash, 0}
	keyNamesToStroke["c-["] = KeyStroke{tcell.KeyCtrlRightSq, 0}
	keyNamesToStroke["c-^"] = KeyStroke{tcell.KeyCtrlCarat, 0}
	keyNamesToStroke["c-_"] = KeyStroke{tcell.KeyCtrlUnderscore, 0}
	keyNamesToStroke["nul"] = KeyStroke{tcell.KeyNUL, 0}
	keyNamesToStroke["soh"] = KeyStroke{tcell.KeySOH, 0}
	keyNamesToStroke["stx"] = KeyStroke{tcell.KeySTX, 0}
	keyNamesToStroke["etx"] = KeyStroke{tcell.KeyETX, 0}
	keyNamesToStroke["eot"] = KeyStroke{tcell.KeyEOT, 0}
	keyNamesToStroke["enq"] = KeyStroke{tcell.KeyENQ, 0}
	keyNamesToStroke["ack"] = KeyStroke{tcell.KeyACK, 0}
	keyNamesToStroke["bel"] = KeyStroke{tcell.KeyBEL, 0}
	keyNamesToStroke["bs"] = KeyStroke{tcell.KeyBS, 0}
	keyNamesToStroke["tab"] = KeyStroke{tcell.KeyTAB, 0}
	keyNamesToStroke["lf"] = KeyStroke{tcell.KeyLF, 0}
	keyNamesToStroke["vt"] = KeyStroke{tcell.KeyVT, 0}
	keyNamesToStroke["ff"] = KeyStroke{tcell.KeyFF, 0}
	keyNamesToStroke["cr"] = KeyStroke{tcell.KeyCR, 0}
	keyNamesToStroke["so"] = KeyStroke{tcell.KeySO, 0}
	keyNamesToStroke["si"] = KeyStroke{tcell.KeySI, 0}
	keyNamesToStroke["dle"] = KeyStroke{tcell.KeyDLE, 0}
	keyNamesToStroke["dc1"] = KeyStroke{tcell.KeyDC1, 0}
	keyNamesToStroke["dc2"] = KeyStroke{tcell.KeyDC2, 0}
	keyNamesToStroke["dc3"] = KeyStroke{tcell.KeyDC3, 0}
	keyNamesToStroke["dc4"] = KeyStroke{tcell.KeyDC4, 0}
	keyNamesToStroke["nak"] = KeyStroke{tcell.KeyNAK, 0}
	keyNamesToStroke["syn"] = KeyStroke{tcell.KeySYN, 0}
	keyNamesToStroke["etb"] = KeyStroke{tcell.KeyETB, 0}
	keyNamesToStroke["can"] = KeyStroke{tcell.KeyCAN, 0}
	keyNamesToStroke["em"] = KeyStroke{tcell.KeyEM, 0}
	keyNamesToStroke["sub"] = KeyStroke{tcell.KeySUB, 0}
	keyNamesToStroke["esc"] = KeyStroke{tcell.KeyESC, 0}
	keyNamesToStroke["fs"] = KeyStroke{tcell.KeyFS, 0}
	keyNamesToStroke["gs"] = KeyStroke{tcell.KeyGS, 0}
	keyNamesToStroke["rs"] = KeyStroke{tcell.KeyRS, 0}
	keyNamesToStroke["us"] = KeyStroke{tcell.KeyUS, 0}
	keyNamesToStroke["del"] = KeyStroke{tcell.KeyDEL, 0}

	keyNames = map[tcell.Key]string{
	tcell.KeyEnter:          "enter",
	tcell.KeyBackspace:      "backspace",
	tcell.KeyTab:            "tab",
	tcell.KeyBacktab:        "backtab",
	tcell.KeyEsc:            "esc",
	tcell.KeyBackspace2:     "backspace2",
	tcell.KeyDelete:         "delete",
	tcell.KeyInsert:         "insert",
	tcell.KeyUp:             "up",
	tcell.KeyDown:           "down",
	tcell.KeyLeft:           "left",
	tcell.KeyRight:          "right",
	tcell.KeyHome:           "home",
	tcell.KeyEnd:            "end",
	tcell.KeyUpLeft:         "upleft",
	tcell.KeyUpRight:        "upright",
	tcell.KeyDownLeft:       "downleft",
	tcell.KeyDownRight:      "downright",
	tcell.KeyCenter:         "center",
	tcell.KeyPgDn:           "pgdn",
	tcell.KeyPgUp:           "pgup",
	tcell.KeyClear:          "clear",
	tcell.KeyExit:           "exit",
	tcell.KeyCancel:         "cancel",
	tcell.KeyPause:          "pause",
	tcell.KeyPrint:          "print",
	tcell.KeyF1:             "f1",
	tcell.KeyF2:             "f2",
	tcell.KeyF3:             "f3",
	tcell.KeyF4:             "f4",
	tcell.KeyF5:             "f5",
	tcell.KeyF6:             "f6",
	tcell.KeyF7:             "f7",
	tcell.KeyF8:             "f8",
	tcell.KeyF9:             "f9",
	tcell.KeyF10:            "f10",
	tcell.KeyF11:            "f11",
	tcell.KeyF12:            "f12",
	tcell.KeyF13:            "f13",
	tcell.KeyF14:            "f14",
	tcell.KeyF15:            "f15",
	tcell.KeyF16:            "f16",
	tcell.KeyF17:            "f17",
	tcell.KeyF18:            "f18",
	tcell.KeyF19:            "f19",
	tcell.KeyF20:            "f20",
	tcell.KeyF21:            "f21",
	tcell.KeyF22:            "f22",
	tcell.KeyF23:            "f23",
	tcell.KeyF24:            "f24",
	tcell.KeyF25:            "f25",
	tcell.KeyF26:            "f26",
	tcell.KeyF27:            "f27",
	tcell.KeyF28:            "f28",
	tcell.KeyF29:            "f29",
	tcell.KeyF30:            "f30",
	tcell.KeyF31:            "f31",
	tcell.KeyF32:            "f32",
	tcell.KeyF33:            "f33",
	tcell.KeyF34:            "f34",
	tcell.KeyF35:            "f35",
	tcell.KeyF36:            "f36",
	tcell.KeyF37:            "f37",
	tcell.KeyF38:            "f38",
	tcell.KeyF39:            "f39",
	tcell.KeyF40:            "f40",
	tcell.KeyF41:            "f41",
	tcell.KeyF42:            "f42",
	tcell.KeyF43:            "f43",
	tcell.KeyF44:            "f44",
	tcell.KeyF45:            "f45",
	tcell.KeyF46:            "f46",
	tcell.KeyF47:            "f47",
	tcell.KeyF48:            "f48",
	tcell.KeyF49:            "f49",
	tcell.KeyF50:            "f50",
	tcell.KeyF51:            "f51",
	tcell.KeyF52:            "f52",
	tcell.KeyF53:            "f53",
	tcell.KeyF54:            "f54",
	tcell.KeyF55:            "f55",
	tcell.KeyF56:            "f56",
	tcell.KeyF57:            "f57",
	tcell.KeyF58:            "f58",
	tcell.KeyF59:            "f59",
	tcell.KeyF60:            "f60",
	tcell.KeyF61:            "f61",
	tcell.KeyF62:            "f62",
	tcell.KeyF63:            "f63",
	tcell.KeyF64:            "f64",
	tcell.KeyCtrlA:          "c-a",
	tcell.KeyCtrlB:          "c-b",
	tcell.KeyCtrlC:          "c-c",
	tcell.KeyCtrlD:          "c-d",
	tcell.KeyCtrlE:          "c-e",
	tcell.KeyCtrlF:          "c-f",
	tcell.KeyCtrlG:          "c-g",
	tcell.KeyCtrlJ:          "c-j",
	tcell.KeyCtrlK:          "c-k",
	tcell.KeyCtrlL:          "c-l",
	tcell.KeyCtrlN:          "c-n",
	tcell.KeyCtrlO:          "c-o",
	tcell.KeyCtrlP:          "c-p",
	tcell.KeyCtrlQ:          "c-q",
	tcell.KeyCtrlR:          "c-r",
	tcell.KeyCtrlS:          "c-s",
	tcell.KeyCtrlT:          "c-t",
	tcell.KeyCtrlU:          "c-u",
	tcell.KeyCtrlV:          "c-v",
	tcell.KeyCtrlW:          "c-w",
	tcell.KeyCtrlX:          "c-x",
	tcell.KeyCtrlY:          "c-y",
	tcell.KeyCtrlZ:          "c-z",
	tcell.KeyCtrlSpace:      "c-space",
	tcell.KeyCtrlUnderscore: "c-_",
	tcell.KeyCtrlRightSq:    "c-]",
	tcell.KeyCtrlBackslash:  "c-\\",
	tcell.KeyCtrlCarat:      "c-^",
}
}
