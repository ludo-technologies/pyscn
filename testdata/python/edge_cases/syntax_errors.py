# Files with syntax errors for testing error handling
# Each error is in a comment block to show what we're testing

# Missing colon
# if x > 0
#     print("positive")

# Unclosed parenthesis
# result = (1 + 2

# Unclosed string
# text = "unclosed string

# Invalid indentation
# def function():
# print("bad indent")

# Unexpected indent
#     unexpected_indent = True

# Mixed tabs and spaces (when strict)
# def mixed():
# 	print("tab")
#     print("spaces")

# Invalid syntax
# def 123invalid():
#     pass

# Missing def keyword
# function_name():
#     pass

# Invalid decorator
# @
# def decorated():
#     pass

# Incomplete statement
# x =

# Invalid assignment
# 5 = x

# Break outside loop
# break

# Continue outside loop
# continue

# Return outside function
# return 42

# Yield outside function
# yield 1

# Invalid operator
# x === 5

# Unclosed bracket
# list = [1, 2, 3

# Unclosed brace
# dict = {"key": "value"

# Invalid lambda
# lambda x, y:

# Invalid comprehension
# [x for in range(10)]

# Missing except
# try:
#     risky()

# Empty except
# try:
#     risky()
# except:

# Invalid import
# from import something

# Circular import (needs multiple files)
# from .syntax_errors import *

# Note: This file intentionally contains syntax errors in comments
# It's used to test the parser's error handling capabilities