"""
E2E tests for REST Monitoring endpoints using pytest-bdd.
"""

from pytest_bdd import scenarios

# Import step definitions
from step_definitions import common_steps, monitoring_steps

# Load scenarios from feature file
scenarios("features/rest/monitoring.feature")
