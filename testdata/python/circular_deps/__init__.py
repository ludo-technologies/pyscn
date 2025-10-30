"""Test data package for circular dependency detection.

This package contains multiple circular dependency patterns:
1. Simple 2-module cycle: module_a <-> module_b
2. 3-module cycle: user_service -> auth_service -> notification_service -> user_service
3. Core infrastructure cycle: database <-> cache <-> logger
4. Larger chain cycle: controller -> service -> repository -> controller

The 'database', 'cache', and 'logger' modules form a core infrastructure
that appears in multiple dependency chains.
"""

__all__ = [
    'module_a',
    'module_b',
    'user_service',
    'auth_service',
    'notification_service',
    'database',
    'cache',
    'logger',
    'controller',
    'service',
    'repository',
]
