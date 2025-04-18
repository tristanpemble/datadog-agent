import unittest

from tasks.libs.testing.flakes import (
    consolidate_flaky_failures,
    get_child_test_in_list,
    get_tests_family,
    get_tests_family_if_failing_tests,
    is_known_flaky_test,
    is_strict_child,
)


class TestGetTestParents(unittest.TestCase):
    def test_get_tests_parents(self):
        parents = get_tests_family(["TestEKSSuite/TestCPU/TestCPUUtilization", "TestKindSuite/TestKind"])
        self.assertEqual(
            parents,
            {
                "TestEKSSuite",
                "TestEKSSuite/TestCPU",
                "TestEKSSuite/TestCPU/TestCPUUtilization",
                "TestKindSuite",
                "TestKindSuite/TestKind",
            },
        )

    def test_get_test_parents_empty(self):
        parents = get_tests_family([])
        self.assertEqual(
            parents,
            set(),
        )

    def test_get_test_parents_failing_no_failing_tests(self):
        parents = get_tests_family_if_failing_tests(["TestEKSSuite/TestCPU/TestCPUUtilization"], set())
        self.assertEqual(
            parents,
            set(),
        )

    def test_get_test_parents_failing_all_failing_tests(self):
        parents = get_tests_family_if_failing_tests(
            ["TestEKSSuite/TestCPU/TestCPUUtilization", "TestKindSuite/TestCPU"],
            {"TestKindSuite/TestCPU", "TestEKSSuite/TestCPU/TestCPUUtilization"},
        )
        self.assertEqual(
            parents,
            {
                "TestEKSSuite",
                "TestEKSSuite/TestCPU",
                "TestEKSSuite/TestCPU/TestCPUUtilization",
                "TestKindSuite",
                "TestKindSuite/TestCPU",
            },
        )

    def test_get_test_parents_failing_some_failing_tests(self):
        parents = get_tests_family_if_failing_tests(
            ["TestEKSSuite/TestCPU/TestCPUUtilization", "TestKindSuite/TestCPU"], {"TestKindSuite/TestCPU"}
        )
        self.assertEqual(
            parents,
            {
                "TestKindSuite",
                "TestKindSuite/TestCPU",
            },
        )


class TestIsKnownFlake(unittest.TestCase):
    def test_known_flake(self):
        is_known_flaky = is_known_flaky_test(
            "TestEKSSuite/mario", {"TestEKSSuite/mario"}, {"TestEKSSuite", "TestEKSSuite/mario"}
        )
        self.assertTrue(is_known_flaky)

    def test_known_flake_parent_failing(self):
        is_known_flaky = is_known_flaky_test(
            "TestEKSSuite", {"TestEKSSuite/mario"}, {"TestEKSSuite", "TestEKSSuite/mario"}
        )
        self.assertTrue(is_known_flaky)

    def test_known_flake_parent_failing_2(self):
        is_known_flaky = is_known_flaky_test(
            "TestEKSSuite/mario",
            {"TestEKSSuite/mario/luigi"},
            {"TestEKSSuite", "TestEKSSuite/mario", "TestEKSSuite/mario/luigi"},
        )
        self.assertTrue(is_known_flaky)

    def test_not_known_flake(self):
        is_known_flaky = is_known_flaky_test(
            "TestEKSSuite/luigi", {"TestEKSSuite/mario"}, {"TestEKSSuite", "TestEKSSuite/mario"}
        )
        self.assertFalse(is_known_flaky)

    def test_not_known_flake_ambiguous_start(self):
        is_known_flaky = is_known_flaky_test("TestEKSSuiteVM/mario", {"TestEKSSuite/mario"}, {"TestEKSSuite"})
        self.assertFalse(is_known_flaky)

    def test_not_known_flake_ambiguous_start_2(self):
        is_known_flaky = is_known_flaky_test("TestEKSSuite/mario", {"TestEKSSuiteVM/mario"}, {"TestEKSSuiteVM"})
        self.assertFalse(is_known_flaky)


class TestConsolidateFlakeFailures(unittest.TestCase):
    def test_one_known_flaky_failure(self):
        flaky_failures = consolidate_flaky_failures({"TestEKSSuite/Mario"}, {"TestEKSSuite/Mario"})
        self.assertEqual(
            flaky_failures,
            {"TestEKSSuite/Mario"},
        )

    def test_one_failure_parent_of_flaky_failure(self):
        flaky_failures = consolidate_flaky_failures({"TestEKSSuite/Mario"}, {"TestEKSSuite", "TestEKSSuite/Mario"})
        self.assertEqual(
            flaky_failures,
            {"TestEKSSuite/Mario", "TestEKSSuite"},
        )

    def test_one_failure_parent_of_non_flaky_failure(self):
        flaky_failures = consolidate_flaky_failures({"TestEKSSuite/Mario"}, {"TestEKSSuite", "TestEKSSuite/Luigi"})
        self.assertEqual(
            flaky_failures,
            {"TestEKSSuite/Mario"},
        )

    def test_recursively_flaky_failures(self):
        flaky_failures = consolidate_flaky_failures(
            {"TestEKSSuite/Mario/Luigi/Wario", "TestEKSSuite/Mario/Luigi/Waluigi"},
            {
                "TestEKSSuite/Mario",
                "TestEKSSuite/Mario/Luigi",
                "TestEKSSuite/Mario/Luigi/Wario",
                "TestEKSSuite/Mario/Luigi/Waluigi",
                "TestKindSuite/Mario",
            },
        )
        self.assertEqual(
            flaky_failures,
            {
                "TestEKSSuite/Mario",
                "TestEKSSuite/Mario/Luigi",
                "TestEKSSuite/Mario/Luigi/Wario",
                "TestEKSSuite/Mario/Luigi/Waluigi",
            },
        )

    def test_recursively_one__non_flaky_failures(self):
        flaky_failures = consolidate_flaky_failures(
            {"TestEKSSuite/Mario/Luigi/Wario", "TestEKSSuite/Mario/Luigi/Waluigi"},
            {
                "TestEKSSuite/Mario",
                "TestEKSSuite/Mario/Luigi",
                "TestEKSSuite/Mario/Luigi/Wario",
                "TestEKSSuite/Mario/Luigi/Waluigi",
                "TestEKSSuite/Mario/Luigi/Yoshi",
                "TestKindSuite/Mario",
            },
        )
        self.assertEqual(
            flaky_failures,
            {"TestEKSSuite/Mario/Luigi/Wario", "TestEKSSuite/Mario/Luigi/Waluigi"},
        )


class TestIsChild(unittest.TestCase):
    def test_is_child(self):
        self.assertTrue(is_strict_child("TestEKSSuite", "TestEKSSuite/TestCPU"))
        self.assertTrue(is_strict_child("TestEKSSuite/TestCPU", "TestEKSSuite/TestCPU/TestCPUUtilization"))

    def test_is_not_child(self):
        self.assertFalse(is_strict_child("TestEKSSuite", "TestKindSuite/TestCPU/TestToto/TestNario"))
        self.assertFalse(is_strict_child("TestEKSSuite/TestCPU", "TestKindSuite/TestCPU/TestCPUUtilization"))


class TestChildInList(unittest.TestCase):
    def test_child_in_list(self):
        children = get_child_test_in_list(
            "TestEKSSuite/TestCPU",
            [
                "TestEKSSuite/TestCPU/TestCPUUtilization",
                "TestEKSSuite/TestCPU",
                "TestKindSuite/TestCPU",
                "TestEKSSuite/TestCPU/TestCPUUtilization/Toto",
            ],
        )
        self.assertEqual(
            children,
            ["TestEKSSuite/TestCPU/TestCPUUtilization", "TestEKSSuite/TestCPU/TestCPUUtilization/Toto"],
        )
