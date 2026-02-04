"""Test file for constructor over-injection anti-pattern."""

class GoodService:
    """A class with acceptable number of constructor parameters."""

    def __init__(self, repo, logger, config):
        self.repo = repo
        self.logger = logger
        self.config = config


class BadService:
    """A class with too many constructor parameters - DI anti-pattern."""

    def __init__(self, user_repo, order_repo, product_repo,
                 payment_service, notification_service, email_service,
                 cache, logger, config):
        self.user_repo = user_repo
        self.order_repo = order_repo
        self.product_repo = product_repo
        self.payment_service = payment_service
        self.notification_service = notification_service
        self.email_service = email_service
        self.cache = cache
        self.logger = logger
        self.config = config


class ExactlyAtThreshold:
    """A class with exactly 5 parameters (at threshold)."""

    def __init__(self, repo1, repo2, service1, service2, logger):
        self.repo1 = repo1
        self.repo2 = repo2
        self.service1 = service1
        self.service2 = service2
        self.logger = logger


class JustOverThreshold:
    """A class with 6 parameters (just over threshold)."""

    def __init__(self, repo1, repo2, service1, service2, logger, config):
        self.repo1 = repo1
        self.repo2 = repo2
        self.service1 = service1
        self.service2 = service2
        self.logger = logger
        self.config = config
