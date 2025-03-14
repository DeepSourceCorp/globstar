def lambda_handler1(event, context):
	eventarg = event['input']
	html3 = "<div> {} </div>"
	html1 = f"<div>{event['input']}</div>"
	html2 = "<div> %s </div".format(eventarg)	
	interm3 = html3.format(eventarg)
	foo = {
		# <no-error>
		"data": event['foo']
	}
	bar(foo)

	result = {
		"statusCode": 200,
		# <expect-error>
		"body": html1,
		"headers": {
			"Content-Type": "text/html"
		}
	}

	result = {
		"statusCode": 200,
		# <expect-error>
		"body": html2,
		"headers": {
			"Content-Type": "text/html"
		}
	}

	result = {
		"statusCode": 200,
		# <expect-error>
		"body": eventarg,
		"headers": {
			"Content-Type": "text/html"
		}
	}

	result = {
		"statusCode": 200,
		# <expect-error>
		"body": interm3,
		"headers": {
			"Content-Type": "text/html"
		}
	}

	result = {
		"statusCode": 200,
		# <expect-error>
		"body": event['url'],
		"headers": {
			"Content-Type": "text/html"
		}
	}
	return result


def handler(event, context):
	eventarg = event['input']
	html = "<div> %s </div>"
	interm = html.format(eventarg)

	return {
		"statusCode": 200,
		# <expect-error>
		"body": interm,
		"headers": {
			"Content-Type": "text/html"
		}
	}