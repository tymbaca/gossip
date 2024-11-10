BAD SHIT:
- created new node (p1)
- added it's address to random node (p2)
- p2 has [p1, p2, p3, p4 | 12:20]
- p4 -> p3 [p2, p3, p4 | 12:21]
- p3 -> p2 [p2, p3, p4 | 12:22] -- p1 GONE

SOLVED

------------------------------------------

- If I have no new info - I will decrease me rate

States:
- Passive: slowly pings other nodes
- Active: got update and rapidely pings other nodes. After one iteration (for 
  range n.peers) switches to Passive state


------------------------------------------

TODO: What if the node "1" is removed and marked as removed in all peer lists, but then node "1" added to cluster again?

------------------------------------------

FIX: dangling removed marks

1. n1 is removed, mark
2. after RemoveTTL i delete n1 from my list
3. some dumbass thinks that n1 is alive and says that to me, now i have n1 alive in my list
4. then someone thinks that n1 is removed and says that to me, now i have n1 marked 
   as removed in my list
5. i ~immediately delete it from my list, bc it's TTL is already gone
6. jmp 3
