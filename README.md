# lib-secs2-hsms-go

An [SECS-II](https://en.wikipedia.org/wiki/SECS-II)/[HSMS](https://en.wikipedia.org/wiki/High-Speed_SECS_Message_Services) library written in go.

Users might need some knowledge on [Go](https://golang.org/), [SEMI Standards](https://en.wikipedia.org/wiki/SEMI#SEMI_standards), and [SECS Message Language (SML)](https://www.peergroup.com/expertise/resources/secs-message-language/) to use this library.

## Usage

1. Install the library

    ```bash
    go get github.com/wolimst/lib-secs2-hsms-go
    ```

2. Import the packages and use them

    Example: using the SECS-II parser

    ```go
    import (
        "github.com/wolimst/lib-secs2-hsms-go/pkg/ast"
        "github.com/wolimst/lib-secs2-hsms-go/pkg/parser/sml"
    )

    func main() {
        messages, errors, warnings := sml.Parse("S1F1 W\n.")
        // ...
    }
    ```

## Features

  1. [Object representation of SECS-II/HSMS Message](#object-representation-of-secs-iihsms-message)
  2. [SML Parser](#sml-parser)
  3. [HSMS Parser](#hsms-parser)

## Object representation of SECS-II/HSMS Message

SECS-II/HSMS messages can be represented using the objects implemented in this library.

The message and data item objects are implemented as following structure.  
Multi-byte string data type such as JIS-8 is not supported currently.

```text
HSMSMessage (Interface)
├── DataMessage
└── ControlMessage

ItemNode (Interface)
├── ASCIINode
├── BinaryNode
├── BooleanNode
├── FloatNode
├── IntNode
├── ListNode
└── UintNode
```

## SML Parser

Parse SML format input string into `DataMessage` object.

### Additional SML syntax

This library extends the [default SML syntax](https://www.peergroup.com/expertise/resources/secs-message-language/), and support following additional syntax.  
These additional syntax are optional; the default syntax can be parsed as well.

1. Message direction (`H->E`, `H<-E` or `H<->E`) can be specified after the wait bit.

    Example:

    ```text
    S1F1 W H->E
    .
    ```

2. Message name can be specified after the message direction, or the wait bit if the message direction is not specified.  
Message name can be any unicode characters except the whitespace characters.

    Example:

    ```text
    S1F1 W H->E AreYouThere?
    .
    ```

3. Line comment  
Any text between `//` and the end of the line is ignored by the parser.
`//` in ASCII quoted string is not a comment.

    Example:

    ```text
    S1F13 W // This is comment
    <L[2]
      <A MDLN>
      <A "1.0.0 // This is not a comment">
    >
    .
    ```

4. Arbitrary data item size

    Example:

    ```text
    S5F11 W
    <L[4]
      <A[32] TIMESTAMP>  // size should be exactly 32 (SML default syntax)
      <A[5..20] EXID>    // size should be in range of 5..20 (SML default syntax)
      <A[..5] EXTYPE>    // size lower limit is 0, upper limit is 5
      <A[5..] EXMESSAGE> // size lower limit is 5, upper limit is not specified
    >
    .    
    ```

## HSMS Parser

Parse HSMS byte sequence into `DataMessage` or `ControlMessage` object.

Example:  
byte sequence `00 00 00 0A FF FF 00 00 00 05 FF FF FF FF` will be parsed to a `ControlMessage` that represent `linktest.req`.
