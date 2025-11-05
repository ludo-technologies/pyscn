"""Module A - Part of a simple 2-module circular dependency."""

import module_b


class ServiceA:
    """Service A that depends on Service B."""

    def __init__(self):
        self.service_b = None

    def use_service_b(self):
        """Use functionality from Service B."""
        if self.service_b is None:
            self.service_b = module_b.ServiceB()
        return self.service_b.process_data("data from A")

    def process_request(self, request):
        """Process a request using both services."""
        result = self.use_service_b()
        return f"ServiceA processed: {request} with {result}"


def helper_function_a():
    """Helper function in module A."""
    return "Helper A"
