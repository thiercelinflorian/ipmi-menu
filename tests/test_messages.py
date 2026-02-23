from __future__ import annotations

from ipmi_menu.config.messages import load_messages, get_available_languages


class TestLoadMessages:
    def test_load_french(self):
        msg = load_messages("fr")
        assert msg.lang == "fr"
        assert "BMC" in msg.t("prompts.bmc_ip")

    def test_load_english(self):
        msg = load_messages("en")
        assert msg.lang == "en"
        assert "BMC" in msg.t("prompts.bmc_ip")

    def test_missing_key_returns_key(self):
        msg = load_messages("en")
        assert msg.t("nonexistent.key") == "nonexistent.key"


class TestTranslation:
    def test_formatting(self):
        msg = load_messages("en")
        result = msg.t("info.hw_detected", vendor="dell", mfg="Dell Inc.", product="R640")
        assert "dell" in result
        assert "Dell Inc." in result
        assert "R640" in result

    def test_format_missing_key_no_crash(self):
        msg = load_messages("en")
        # Should not raise even if format keys don't match
        result = msg.t("info.hw_detected")
        assert isinstance(result, str)


class TestAvailableLanguages:
    def test_languages_list(self):
        langs = get_available_languages()
        codes = [code for code, _ in langs]
        assert "fr" in codes
        assert "en" in codes

    def test_all_keys_consistent(self):
        fr = load_messages("fr")
        en = load_messages("en")
        fr_keys = set(fr.data.keys())
        en_keys = set(en.data.keys())
        assert fr_keys == en_keys, f"Missing keys: FR-EN={fr_keys - en_keys}, EN-FR={en_keys - fr_keys}"
