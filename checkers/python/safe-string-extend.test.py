from django.utils.safestring import SafeString, SafeData, SafeText

# <expect-error>
class IWantToBypassEscaping(SafeString):
    def __init__(self):
        super().__init__()

# <expect-error>
class IWantToBypassEscaping2(SafeText):
    def __init__(self):
        super().__init__()

# <expect-error>
class IWantToBypassEscaping3(SafeData):
    def __init__(self):
        super().__init__()

# <no-error>
class SomethingElse(str):
    def __init__(self):
        super().__init__()
