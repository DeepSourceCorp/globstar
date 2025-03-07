def bad_default(arg, container=[]):  # <expect-error>
    container.append(arg)
    return container

# This function uses None as the default and initializes the list inside the function.
def good_default(arg, container=None):  # <no-error>
    if container is None:
        container = []
    container.append(arg)
    return container