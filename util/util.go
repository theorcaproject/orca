package util

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "strings"
)


func HighlightBytePosition(f io.Reader, pos int64) (line, col int, highlight string) {
    line = 1
    br := bufio.NewReader(f)
    lastLine := ""
    thisLine := new(bytes.Buffer)
    for n := int64(0); n < pos; n++ {
        b, err := br.ReadByte()
        if err != nil {
            break
        }
        if b == '\n' {
            lastLine = thisLine.String()
            thisLine.Reset()
            line++
            col = 1
        } else {
            col++
            thisLine.WriteByte(b)
        }
    }
    if line > 1 {
        highlight += fmt.Sprintf("%5d: %s\n", line-1, lastLine)
    }
    highlight += fmt.Sprintf("%5d: %s\n", line, thisLine.String())
    highlight += fmt.Sprintf("%s^\n", strings.Repeat(" ", col+5))
    return
}

