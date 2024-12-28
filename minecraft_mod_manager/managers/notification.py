"""Discord notifications and player warnings."""

from datetime import datetime, timezone
import logging
import time
from typing import Union

import requests

from ..config.config import Config
from ..utils.constants import (
    DISCORD_ERROR_COLOR,
    DISCORD_FOOTER_TEXT,
    DISCORD_MAX_LENGTH,
    DISCORD_SUCCESS_COLOR,
    WARNING_SLEEP_MINUTES,
    WARNING_SLEEP_SECONDS,
)

class NotificationManager:
    """Handles Discord notifications and player warnings."""
    
    def __init__(self, config: Config, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.webhook_url = config['notifications']['discord_webhook']
    
    def send_discord_notification(self, title: str, message: str, is_error: bool = False) -> None:
        """
        Send formatted notification to Discord webhook.
        
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
            payload = {
                "embeds": [{
                    "title": title,
                    "description": message,
                    "color": DISCORD_ERROR_COLOR if is_error else DISCORD_SUCCESS_COLOR,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "footer": {"text": DISCORD_FOOTER_TEXT}
                }]
            }
            
            # Send webhook request
            response = requests.post(
                self.webhook_url,
                json=payload,
                headers={'Content-Type': 'application/json'},
                timeout=10
            )
            
            if response.status_code not in (200, 204):
                raise RuntimeError(f"Discord API returned status {response.status_code}")
                
        except requests.Timeout:
            self.logger.error("Discord notification timed out")
        except requests.RequestException as e:
            self.logger.error(f"Discord API request failed: {str(e)}")
        except Exception as e:
            self.logger.error(f"Failed to send Discord notification: {str(e)}")
    
    def warn_players(self) -> None:
        """Send countdown warnings to players before maintenance."""
        warnings = self.config['maintenance']['warning_intervals']
        if not warnings:
            self.logger.warning("No warning intervals configured - skipping player warnings")
            return
        
        try:
            for warning in warnings:
                time_val = warning['time']
                unit = warning['unit'].lower()
                
                if not isinstance(time_val, (int, float)) or time_val <= 0:
                    self.logger.warning(f"Invalid warning time value: {time_val}")
                    continue
                
                if unit not in ('minutes', 'seconds'):
                    self.logger.warning(f"Invalid warning time unit: {unit}")
                    continue
                
                self._send_warning_message(time_val, unit)
                sleep_time = (WARNING_SLEEP_MINUTES if unit == 'minutes' 
                            else WARNING_SLEEP_SECONDS)
                time.sleep(sleep_time)
                
            self._send_final_warning()
            
        except Exception as e:
            self.logger.error(f"Failed to send player warnings: {str(e)}")
    
    def _send_warning_message(self, time_val: Union[int, float], unit: str) -> None:
        """Send a single warning message to players."""
        try:
            self.send_server_message(
                f"§c[WARNING] Server maintenance in {time_val} {unit}!"
            )
        except Exception as e:
            self.logger.error(f"Failed to send warning message: {str(e)}")
    
    def _send_final_warning(self) -> None:
        """Send final warning message before maintenance."""
        try:
            self.send_server_message("§c[WARNING] Starting maintenance now!")
        except Exception as e:
            self.logger.error(f"Failed to send final warning: {str(e)}")
    
    def send_server_message(self, message: str) -> None:
        """Send in-game message (requires RCON support)."""
        self.logger.warning("Server messages not implemented - no RCON support") 