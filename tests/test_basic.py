"""Basic tests for minecraft-mod-manager."""

from minecraft_mod_manager import __version__


def test_version():
    """Test that version is a string."""
    assert isinstance(__version__, str)
    assert len(__version__) > 0 