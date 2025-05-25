"""
E2E tests for Connect-Go Health Service using pytest-bdd.
"""

from pytest_bdd import scenarios

# Import step definitions
from step_definitions import common_steps, connect_steps

# Load scenarios from feature file
scenarios("features/connect/health_service.feature")
