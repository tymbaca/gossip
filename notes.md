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
