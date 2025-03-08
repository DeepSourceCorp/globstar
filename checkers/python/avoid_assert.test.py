def divide(a, b):
    # <expect-error>
    assert b != 0
    return a / b

# <expect-error>
assert 1 == 1
# <expect-error>
assert 1 == 2
