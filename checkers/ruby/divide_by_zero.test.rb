# <expect-error>
a = 1 / 0

# <expect-error>
b = a / 0

# safe
d = a / 2

# Safe
puts a/2

# <expect-error>
puts a/0
 
# <expect-error> c is being used as divisor
c = 0

puts a/c 

d = a/c