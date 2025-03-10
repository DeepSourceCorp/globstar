from urllib.request import OpenerDirector
from urllib.request import Request
from urllib.request import urlopen
from urllib.request import URLopener
from urllib.request import urlretrieve



def test1():
    od = OpenerDirector()
    # <expect-error>
    od.open("ftp://example.com")

def test1_ok():
    od = OpenerDirector()
    # <no-error>
    od.open("sftp://example.com")

def test2():
    od = OpenerDirector()
    url2 = "ftp://example.com"
    # <expect-error>
    od.open(url2)

def test2_ok():
    od = OpenerDirector()
    # <no-error>
    url2_ok = "sftp://example.com"
    od.open(url2_ok)

def test3():
    # <expect-error>
    OpenerDirector().open("ftp://example.com")

def test3_ok():
    # <no-error>
    OpenerDirector().open("sftp://example.com")

def test4():
    url4 = "ftp://example.com"
    # <expect-error>
    OpenerDirector().open(url4)

def test4_ok():
    # <no-error>
    url4_ok = "sftp://example.com"
    OpenerDirector().open(url4_ok)

def test5(url5 = "ftp://example.com"):
    # <expect-error>
    OpenerDirector().open(url5)

def test5_ok(url5_ok = "sftp://example.com"):
    # <no-error>
    OpenerDirector().open(url5_ok)

def test6(url6 = "ftp://example.com"):
    od = OpenerDirector()
    # <expect-error>
    od.open(url6)

def test6_ok(url6_ok = "sftp://example.com"):
    od = OpenerDirector()
    # <no-error>
    od.open(url6_ok)


def test7():
    # <expect-error>
    Request("ftp://example.com")

def test7_ok():
    # <no-error>
    Request("sftp://example.com")

def test8():
    url8 = "ftp://example.com"
    # <expect-error>
    Request(url8)

def test8_ok():
    # <no-error>
    url8_ok = "sftp://example.com"
    Request(url8_ok)

def test9(url9 = "ftp://example.com"):
    # <expect-error>
    Request(url9)

def test9_ok(url9_ok = "sftp://example.com"):
    # <no-error>
    Request(url9_ok)

def test10():
    # <expect-error>
    urlopen("ftp://example.com")

def test10_ok():
    # <no-error>
    urlopen("sftp://example.com")

def test11():
    url11 = "ftp://example.com"
    # <expect-error>
    urlopen(url11)

def test11_ok():
    url11_ok = "sftp://example.com"
    # <no-error>
    urlopen(url11_ok)

def test12(url12 = "ftp://example.com"):
    # <expect-error>
    urlopen(url12)

def test12_ok(url12_ok = "sftp://example.com"):
    # <no-error>
    urlopen(url12_ok)


def test13():
    od = URLopener()
    # <expect-error>
    od.open("ftp://example.com")

def test13_ok():
    od = URLopener()
    # <no-error>
    od.open("ftps://example.com")

def test14():
    od = URLopener()
    url14 = "ftp://example.com"
    # <expect-error>
    od.open(url14)

def test14_ok():
    od = URLopener()
    # <no-error>
    url14_ok = "ftps://example.com"
    od.open(url14_ok)

def test15():
    # <expect-error>
    URLopener().open("ftp://example.com")

def test15_ok():
    # <no-error>
    URLopener().open("ftps://example.com")

def test16():
    url16 = "ftp://example.com"
    # <expect-error>
    URLopener().open(url16)

def test16_ok():
    # <no-error>
    url16_ok = "ftps://example.com"
    URLopener().open(url16_ok)

def test17(url17 = "ftp://example.com"):
    # <expect-error>
    URLopener().open(url17)

def test17_ok(url17_ok = "ftps://example.com"):
    # <no-error>
    URLopener().open(url17_ok)

def test18(url18 = "ftp://example.com"):
    od = URLopener()
    # <expect-error>
    od.open(url18)

def test18_ok(url18_ok = "ftps://example.com"):
    od = URLopener()
    # <no-error>
    od.open(url18_ok)

def test19():
    od = URLopener()
    # <expect-error>
    od.retrieve("ftp://example.com")

def test19_ok():
    od = URLopener()
    # <no-error>
    od.retrieve("ftps://example.com")

def test20():
    od = URLopener()
    url20 = "ftp://example.com"
    # <expect-error>
    od.retrieve(url20)

def test20_ok():
    od = URLopener()
    # <no-error>
    url20_ok = "ftps://example.com"
    od.retrieve(url20_ok)

def test21():
    # <expect-error>
    URLopener().retrieve("ftp://example.com")

def test21_ok():
    # <no-error>
    URLopener().retrieve("ftps://example.com")

def test22():
    url22 = "ftp://example.com"
    # <expect-error>
    URLopener().retrieve(url22)

def test22_ok():
    # <no-error>
    url22_ok = "ftps://example.com"
    URLopener().retrieve(url22_ok)

def test23(url23 = "ftp://example.com"):
    # <expect-error>
    URLopener().retrieve(url23)

def test23_ok(url23_ok = "ftps://example.com"):
    # <no-error>
    URLopener().retrieve(url23_ok)

def test24(url24 = "ftp://example.com"):
    od = URLopener()
    # <expect-error>
    od.retrieve(url24)

def test24_ok(url24_ok = "ftps://example.com"):
    od = URLopener()
    # <no-error>
    od.retrieve(url24_ok)

def test25():
    # <expect-error>
    urlretrieve("ftp://example.com")

def test25_ok():
    # <no-error>
    urlretrieve("sftp://example.com")

def test26():
    url26 = "ftp://example.com"
    # <expect-error>
    urlretrieve(url26)

def test26_ok():
    # <no-error>
    url26_ok = "sftp://example.com"
    urlretrieve(url26_ok)

def test27(url27 = "ftp://example.com"):
    # <expect-error>
    urlretrieve(url27)

# <no-error>
def test28_ok(url28_ok = "sftp://example.com"):
    urlretrieve(url28_ok)

