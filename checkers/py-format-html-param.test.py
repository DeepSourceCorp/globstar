from django.utils.html import format_html

planet = "world"
markup = "<marquee>" + planet

#############################################################

# <no-error>
print(format_html("hello {}", markup))
# <no-error>
print(format_html("hello {}", "<marquee>world"))
# <no-error>
print(format_html("hello {}", "<marquee>" "world"))
# <no-error>
print(format_html("hello {}", "<marquee>" + "world"))
# <no-error>
print(format_html("hello {}", f"<marquee>{planet}"))
# <no-error>
print(format_html("hello {}", "<marquee>%s" % planet))
# <no-error>
print(format_html("hello {}", "<marquee>{}".format(planet)))
# <no-error>
print(format_html("hello " "{}", "<marquee>world"))
# <no-error>
print(format_html("hello " + "{}", "<marquee>world"))

#############################################################

# <expect-error>
print(format_html("hello %s" % markup))
# <expect-error>
print(format_html(f"hello {markup}"))
# <expect-error>
print(format_html("hello {}".format(markup)))
# <expect-error>
print(format_html("hello %s {}" % markup, markup))
# <expect-error>
print(format_html(f"hello {markup} {{}}", markup))
# <expect-error>
print(format_html("hello {} {{}}".format(markup), markup))
