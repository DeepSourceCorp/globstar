# Test cases for py-no-else-after-return

# Case 1: Violation - simple if/else with return
def foo(x):
    if x > 0:
        return "positive"
    else: # This 'else' is unnecessary
        return "non-positive"

# Case 2: Violation - if/elif/else with return in if
# Note: The current pattern might not be smart enough to catch this if the `elif` is present.
# We will see if the pattern needs adjustment.
def bar(x):
    if x > 10:
        return "large"
    elif x > 5:
        print("medium")
    else: # This 'else' is unnecessary if the first 'if' returns
        return "small"

# Case 3: No violation - if without else
def baz(x):
    if x:
        return True
    print("after if")
    return False

# Case 4: No violation - if/else but no return in the if block
def qux(x):
    if x:
        print("x is true")
    else:
        print("x is false")
    return None

# Case 5: Violation - if with multiple statements and return, then else
def corge(y):
    if y == 1:
        print("one")
        return "uno"
    else: # Unnecessary
        print("not one")
        return "no uno"

# Case 6: No violation - if/elif/else where if and elif don't always return
def grault(z):
    if z > 100:
        if z > 200:
            return "very large"
        print("still large")
    elif z > 50:
        return "medium"
    else:
        return "small" # This else is necessary
