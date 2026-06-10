# Non-clone example: apispec-style field converter (issue #482)
# Shares its control-flow template with field_converter_b.py but documents a
# different field type with entirely different spec keys, so the pair must
# not be reported as a Type-4 clone.


def delimited_list2param(self, field, **kwargs):
    """Document DelimitedList field as a query parameter."""
    ret = {}
    if isinstance(field, DelimitedList):
        if self.openapi_version.major < 3:
            ret["collectionFormat"] = "csv"
        else:
            ret["explode"] = False
            ret["style"] = "form"
    return ret
