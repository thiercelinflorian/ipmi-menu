from __future__ import annotations

from unittest import mock

from ipmi_menu.core.utils import _sanitize_cmd, run_cmd


class TestSanitizeCmd:
    def test_password_masked(self):
        cmd = ["ipmitool", "-H", "1.2.3.4", "-U", "root", "-P", "secret", "mc", "info"]
        result = _sanitize_cmd(cmd)
        assert result[6] == "****"
        assert "secret" not in result

    def test_no_password(self):
        cmd = ["ipmitool", "-H", "1.2.3.4", "-U", "root", "mc", "info"]
        result = _sanitize_cmd(cmd)
        assert result == cmd

    def test_password_at_end(self):
        cmd = ["ipmitool", "-P", "pass"]
        result = _sanitize_cmd(cmd)
        assert result == ["ipmitool", "-P", "****"]


class TestRunCmd:
    def test_success(self):
        rc, out, err = run_cmd(["echo", "hello"], timeout=5)
        assert rc == 0
        assert out == "hello"

    def test_command_not_found(self):
        rc, out, err = run_cmd(["nonexistent_command_xyz"], timeout=5)
        assert rc == 127
        assert err == "command not found"

    def test_timeout(self):
        rc, out, err = run_cmd(["sleep", "10"], timeout=1)
        assert rc == 124
        assert err == "timeout"

    def test_with_mock(self):
        mock_result = mock.MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = "Chassis Power is on\n"
        mock_result.stderr = ""

        with mock.patch("ipmi_menu.core.utils.subprocess.run", return_value=mock_result):
            rc, out, err = run_cmd(["ipmitool", "power", "status"], timeout=10)
            assert rc == 0
            assert out == "Chassis Power is on"
