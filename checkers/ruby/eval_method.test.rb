# Unsafe: Using eval (Avoid this)
params = { 'b' => '1 + 1', 'l' => '2 * 3' }
b = 10

# <expect-error> use of eval
puts eval("fun1", b)  # Undefined method risk

# <expect-error> use of eval
puts eval(params['b'], b)  # Executes user-controlled input (RCE risk)

# <expect-error> use of eval
puts eval(params.dig('l'))  # Executes arbitrary code

# Safe Alternative: Using case statements or explicit logic
def safe_eval(expression)
  allowed_operations = %w[+ - * /]
  tokens = expression.split

  if tokens.size == 3 && allowed_operations.include?(tokens[1])
    num1 = Integer(tokens[0]) rescue nil
    num2 = Integer(tokens[2]) rescue nil

    return num1.send(tokens[1], num2) if num1 && num2
  end

  raise "Unsafe operation detected!"
end

# Using safe alternative instead of eval
begin
  safe_result = safe_eval(params['b'])  # Securely evaluate math expressions
  puts "Safe Evaluated Result: #{safe_result}"
rescue => e
  puts "Error: #{e.message}"
end
