```
% sll-example 
[ PushTail objects ]
one
two
three
four
five
six
seven
eight
nine
ten

[ traverse from tail ]
ten -> nine -> eight -> seven -> six -> five -> four -> three -> two -> one ->

[ traverse from head ]
one -> two -> three -> four -> five -> six -> seven -> eight -> nine -> ten ->

[ read tail 3x ]
ten
ten
ten

[ read top 2 scores ]
Value:nine Score:2
Value:ten Score:5

[ move tail to head ]
Current: ten -> nine -> eight -> seven -> six -> five -> four -> three -> two -> one ->
New: nine -> eight -> seven -> six -> five -> four -> three -> two -> one -> ten ->

[ remove head, tail ]
Current: nine -> eight -> seven -> six -> five -> four -> three -> two -> one -> ten ->
New: eight -> seven -> six -> five -> four -> three -> two -> one ->

[ remove second from last node ]
eight -> six -> five -> four -> three -> two -> one ->

[ read score list ]
one two three four five six eight

```
