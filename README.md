# react4blub

Some experimental attempts to re-implement the React model in "blub" languages.

- [./c](./c) - C programming language. I learned a lot about macros and
  metaprogramming, but couldn't figure out how to handle allocations for the
  `H(comp, props, children...)` return value of a component. Ideally this could
  just be a struct on the stack, but the variable-sized children array poses an
  issue I can't solve with my current (rudimentary) C skills.

- Go (current attempt). This is going much better; I have a prototype
  RenderToString working fine. I'm currently stuck on figuring out/writing the
  reconciliation algorithm. Go generics make this much more possible/appealing
  compared to what was possible in the past!
