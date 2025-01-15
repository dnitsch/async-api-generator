# Internals

![Diagram of the internal flow](./EventCatalog-Internal-Flow.png)

## Lexer

TOKENs are kept to the following types - `token.TokenType`  - not listed as the list is likely to change/grow.

Special consideration will need to be given to files that are not able to contain comments or anything outside of their existing syntax - e.g. .json most commonly containing schemas. 

> these cases a convention will need to be followed where by the name of the message that it is describing must be in the name of the file.

## Parser

Not using an existing parser generator like CGF, BNF, or EBNF is on purposes as the input source will only ever really be composed of parts we care about i.e. `gendoc` markers their beginning and end and what they enclose within them as text.

We'll use the overly simplified Pratt Parser (top down method) as we'll have no need to for expression parsing only statement nodes creation with the associated/helper attributes.

[more...]()

### Generation

Once a flat list of statements (`[]GenDocBlock`) is ready we then need to sort it in order of precedence. Precedence is set based on `category` (shorthand `c`) found on an annotation.

[more...]()

once sorted we need to build an interim tree, as at this point we have no idea how many nodes there will be, it has to be an `n-ary tree`

## Gendoc Tree

The tree looks like the below diagram. 

![](./EventCatalog-ServiceContextTree.png)

Then highlights the order in which it's walked. It is using the __BFS (BreadthFirstSearch) algorithm__ to walk each level and perform the merging of information from all the *leaf*  nodes.

Also worth noting is that it is using an internal indexer for quicker O(n) lookups when performing the sort. The tree is walked multiple times to ensure the orphans are assigned to parents in case they weren't in the tree when it was walked previously.

```mermaid
flowchart TD
    root(0_""\nroot)
    orphaned(0_orphaned)
    any_any[any_any]

    parented(0_parented)
    srv(1_serviceId)
    chan(2_channelId)
    op(3_operationId)
    msg(4_messageId)
    usrv[unsorted srvs]
    umsg[unsorted messages]
    uop[unsorted operations]
    uchan[unsorted channels]
    
    root --> orphaned
    orphaned -.-> |1..n| any_any[any_any]
    root --> parented
    parented --> |1..n| srv 
    srv -.->|1..n| usrv
    srv -->|1..n| chan
    chan -.->|1..n| uchan
    chan -->|1..1| op
    op -.->|1..n| uop
    op -->|1..1| msg
    msg -->|1..n| umsg 
```
