"""Discord notifications and player warnings."""

import logging
from datetime import datetime, timezone
from typing import Any, Dict

import aiohttp

from ..config.config import Config
from ..utils.constants import (
    DISCORD_ERROR_COLOR,
    DISCORD_FOOTER_TEXT,
    DISCORD_MAX_LENGTH,
    DISCORD_SUCCESS_COLOR,
)


class NotificationManager:
    """Handles Discord notifications and player warnings."""

    def __init__(self, config: Config, logger: logging.Logger) -> None:
        """Initialize notification manager."""
        self.config = config
        self.logger = logger
        self.webhook_url = config.notifications.discord_webhook

    async def send_discord_notification(self, title: str, message: str, is_error: bool = False) -> None:
        """Send formatted notification to Discord webhook.

        Args:
            title: Notification title
            message: Notification message
            is_error: Whether this is an error notification (changes color)
        """
        if not self.webhook_url:
            self.logger.debug("Discord notifications disabled - no webhook URL configured")
            return

        try:
            # Truncate message if too long
            if len(message) > DISCORD_MAX_LENGTH:
                message = message[:DISCORD_MAX_LENGTH - 3] + "..."

            # Format Discord embed
            payload: Dict[str, Any] = {
                "embeds": [{
                    "title": title,
                    "description": message,
                    "color": DISCORD_ERROR_COLOR if is_error else DISCORD_SUCCESS_COLOR,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "footer": {"text": DISCORD_FOOTER_TEXT}
                }]
            }

            async with aiohttp.ClientSession() as session:
                async with session.post(self.webhook_url, json=payload) as response:
                    if response.status not in (200, 204):
                        raise RuntimeError(f"Discord API returned status {response.status}")

        except asyncio.TimeoutError:
            self.logger.error("Discord notification timed out")
        except Exception as e:
            self.logger.error(f"Failed to send Discord notification: {str(e)}")

    async def warn_players(self) -> None:
        """Send warning messages to online players about upcoming server restart."""
        try:
            # Format warning message
            warning_msg = self.config.notifications.warning_template.format(
                minutes=self.config.notifications.warning_intervals[0]
            )

            # TODO: Implement actual player warning via server commands
            # For now just log the warning
            self.logger.info(f"Server restart warning: {warning_msg}")
            await self.send_discord_notification(
                "Server Warning",
                warning_msg
            )
        except Exception as e:
            self.logger.error(f"Failed to send player warning: {str(e)}")
