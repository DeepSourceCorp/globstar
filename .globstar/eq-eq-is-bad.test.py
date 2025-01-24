x = 1

# <expect-error>
x == x

foo = 1
if foo:
  # <expect-error>
  foo == foo

assertTrue(foo == foo)  # this is OK
