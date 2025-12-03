# LexMD
*Yes it's more than a lexer, but now the name sounds like my wife's favorite website...*

# Parser
    - create "doc" containing blocks. can be added as tokens are coming in from
    lexer.
    - once block is completed can parse lines
    
# Lexer
    - what to tokenize?
    - need tokens for valid begin and end of blocks
    - tokens for inline elements
    - white space?

# Potential state functions
    - lexBlockStart
    - lexBlockEnd
    - lexLine
    - lexHeading

