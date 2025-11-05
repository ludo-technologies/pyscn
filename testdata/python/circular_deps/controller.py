"""Controller module - Part of larger dependency chain."""

import service
import auth_service


class UserController:
    """HTTP controller for user endpoints."""

    def __init__(self):
        self.service = service.BusinessService()
        self.auth = auth_service.AuthService()

    def handle_request(self, request):
        """Handle incoming HTTP request."""
        if not self.auth.validate_credentials(request.get("user"), request.get("password")):
            return {"error": "Unauthorized"}

        result = self.service.process(request.get("data"))
        return {"result": result}


class ApiController:
    """API controller for external integrations."""

    def handle_api_request(self, api_request):
        """Handle API request."""
        return {"status": "ok"}
