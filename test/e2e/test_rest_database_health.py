"""
E2E tests for REST Database Health endpoints using pytest-bdd.
"""

from pytest_bdd import scenarios

# Import step definitions
from step_definitions import common_steps, database_steps

# Load scenarios from feature file
scenarios("features/rest/database_health.feature")
