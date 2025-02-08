from django.utils.html import (
    conditional_escape, escape, escapejs, format_html, html_safe, json_script,
    linebreaks, smart_urlquote, strip_spaces_between_tags, strip_tags, urlize,
)
from django.utils.safestring import mark_safe

# cf.https://github.com/django/django/blob/76ed1c49f804d409cfc2911a890c78584db3c76e/tests/utils_tests/test_html.py#L204
# <expect-error>
@html_safe
class HtmlClass:
    def __str__(self):
        return "<h1>I'm a html class!</h1>"

# <no-error>
class Boring:
    def __str__(self):
        return "<h1>I will become an html class!</h1>"

# <expect-error>
HtmlBoring = html_safe(Boring)
