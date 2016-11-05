```
% sll-example 
[ PushTail objects ]
one
two
three
four
five

[ traverse from tail ]
five -> four -> three -> two -> one ->

[ traverse from head ]
one -> two -> three -> four -> five ->

[ read tail 3x ]
five
five
five

[ read top 2 scores ]
Value:four Score:2
Value:five Score:5

[ move tail to head ]
Current: five -> four -> three -> two -> one ->
New: four -> three -> two -> one -> five ->

[ remove head, tail ]
Current: four -> three -> two -> one -> five ->
New: three -> two -> one ->

[ remove middle node ]
three -> one ->
```