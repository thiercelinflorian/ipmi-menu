from __future__ import annotations

from ipmi_menu.core.ipmi import looks_like_auth_error, normalize_vendor, parse_kv


class TestParseKv:
    def test_basic(self):
        text = "Manufacturer : Supermicro\nProduct Name : X11\n"
        result = parse_kv(text)
        assert result["manufacturer"] == "Supermicro"
        assert result["product name"] == "X11"

    def test_empty(self):
        assert parse_kv("") == {}

    def test_no_colon(self):
        assert parse_kv("no colon here\nanother line") == {}

    def test_multiple_colons(self):
        result = parse_kv("key : val : extra")
        assert result["key"] == "val : extra"


class TestNormalizeVendor:
    def test_supermicro(self):
        assert normalize_vendor("Supermicro", "X11", "") == "supermicro"

    def test_dell(self):
        assert normalize_vendor("Dell Inc.", "PowerEdge", "") == "dell"

    def test_hp(self):
        assert normalize_vendor("HP", "ProLiant", "") == "hp"

    def test_lenovo(self):
        assert normalize_vendor("Lenovo", "ThinkSystem", "") == "lenovo"

    def test_asrockrack(self):
        assert normalize_vendor("ASRockRack", "EPYC", "") == "asrockrack"

    def test_intel(self):
        assert normalize_vendor("Intel Corporation", "S2600", "") == "intel"

    def test_unknown(self):
        assert normalize_vendor("", "", "") == "unknown"

    def test_fallback(self):
        assert normalize_vendor("Acme Corp", "Model X", "") == "acme corp"


class TestLooksLikeAuthError:
    def test_auth_error(self):
        assert looks_like_auth_error("RAKP 2 HMAC is invalid")

    def test_unauthorized(self):
        assert looks_like_auth_error("Error: Unauthorized access")

    def test_timeout(self):
        assert looks_like_auth_error("Connection timeout occurred")

    def test_normal_output(self):
        assert not looks_like_auth_error("Chassis Power is on")

    def test_empty(self):
        assert not looks_like_auth_error("")

    def test_none(self):
        assert not looks_like_auth_error(None)
