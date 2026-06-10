# Non-clone example: apispec-style field converter (issue #482)
# Shares its control-flow template with field_converter_a.py but documents a
# different field type with entirely different spec keys, so the pair must
# not be reported as a Type-4 clone.


def uploadfield2properties(self, field, **kwargs):
    """Document Upload field properties in the API spec."""
    ret = {}
    if isinstance(field, Upload):
        if self.openapi_version.major < 3:
            ret["type"] = "file"
        else:
            ret["type"] = "string"
            ret["format"] = field.format
    return ret
