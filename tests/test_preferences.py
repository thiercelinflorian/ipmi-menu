from __future__ import annotations

import base64
import json
import os
import stat
from pathlib import Path
from unittest import mock

import pytest

from ipmi_menu.config import preferences


@pytest.fixture
def tmp_prefs(tmp_path):
    """Override preferences file path to use a temp directory."""
    prefs_file = tmp_path / "preferences.json"
    config_dir = tmp_path
    with mock.patch.object(preferences, "PREFERENCES_FILE", prefs_file), \
         mock.patch.object(preferences, "CONFIG_DIR", config_dir):
        yield prefs_file


class TestLoadSave:
    def test_save_and_load(self, tmp_prefs):
        prefs = {"language": "en", "version": 1}
        preferences.save_preferences(prefs)
        loaded = preferences.load_preferences()
        assert loaded["language"] == "en"

    def test_load_missing_file(self, tmp_prefs):
        assert preferences.load_preferences() == {}

    def test_load_corrupt_file(self, tmp_prefs):
        tmp_prefs.write_text("not json{{{")
        assert preferences.load_preferences() == {}


class TestPermissions:
    def test_file_permissions_0600(self, tmp_prefs):
        preferences.save_preferences({"language": "fr", "version": 1})
        mode = stat.S_IMODE(os.stat(tmp_prefs).st_mode)
        assert mode == 0o600


class TestBase64Password:
    def test_password_stored_as_base64(self, tmp_prefs):
        preferences.set_preferred_password("s3cret")
        raw = json.loads(tmp_prefs.read_text())
        encoded = raw["password"]
        assert encoded == base64.b64encode(b"s3cret").decode("ascii")

    def test_password_decoded_on_read(self, tmp_prefs):
        preferences.set_preferred_password("s3cret")
        assert preferences.get_preferred_password() == "s3cret"

    def test_password_none(self, tmp_prefs):
        preferences.set_preferred_password(None)
        assert preferences.get_preferred_password() is None

    def test_clear_password(self, tmp_prefs):
        preferences.set_preferred_password("test")
        preferences.set_preferred_password(None)
        assert preferences.get_preferred_password() is None


class TestMigration:
    def test_migrate_plaintext_password(self, tmp_prefs):
        # Write a v0 (no version) prefs with plaintext password
        tmp_prefs.write_text(json.dumps({"password": "oldpass"}))
        os.chmod(tmp_prefs, 0o644)

        loaded = preferences.load_preferences()
        # After migration, the password should be readable
        pw = preferences.get_preferred_password()
        assert pw == "oldpass"

        # File should now have version key and base64 password
        raw = json.loads(tmp_prefs.read_text())
        assert raw["version"] == 1
        assert raw["password"] == base64.b64encode(b"oldpass").decode("ascii")

    def test_no_double_migration(self, tmp_prefs):
        """Already base64 passwords should not be re-encoded."""
        encoded = base64.b64encode(b"mypass").decode("ascii")
        tmp_prefs.write_text(json.dumps({"password": encoded, "version": 1}))
        os.chmod(tmp_prefs, 0o600)

        pw = preferences.get_preferred_password()
        assert pw == "mypass"
