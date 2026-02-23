from __future__ import annotations

from ipmi_menu.core.updater import _parse_version


class TestParseVersion:
    def test_normal(self):
        assert _parse_version("1.2.3") == (1, 2, 3)

    def test_two_parts(self):
        assert _parse_version("1.2") == (1, 2)

    def test_single(self):
        assert _parse_version("5") == (5,)

    def test_invalid(self):
        assert _parse_version("abc") == (0, 0, 0)

    def test_empty(self):
        assert _parse_version("") == (0, 0, 0)

    def test_extra_parts_ignored(self):
        assert _parse_version("1.2.3.4.5") == (1, 2, 3)


class TestVersionComparison:
    def test_newer(self):
        assert _parse_version("2.0.0") > _parse_version("1.9.9")

    def test_equal(self):
        assert _parse_version("1.0.7") == _parse_version("1.0.7")

    def test_older(self):
        assert _parse_version("1.0.6") < _parse_version("1.0.7")

    def test_minor_bump(self):
        assert _parse_version("1.1.0") > _parse_version("1.0.99")
