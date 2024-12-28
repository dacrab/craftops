"""Command-line interface for Minecraft Mod Manager."""

import argparse
import asyncio
import sys
from enum import Enum, auto
from typing import NoReturn

from .minecraft_mod_manager import MinecraftModManager
from .utils.constants import DEFAULT_CONFIG_PATH


class ServerAction(Enum):
    """Server control actions."""
    START = auto()
    STOP = auto()
    RESTART = auto()
    STATUS = auto()


def parse_args() -> argparse.Namespace:
    """Parse and validate command line arguments."""
    parser = argparse.ArgumentParser(
        description='Minecraft Mod Manager - Server management and mod update tool',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
            Examples:
              %(prog)s --status
              %(prog)s --start
              %(prog)s --auto-update
              %(prog)s --config custom_config.jsonc --auto-update
        """
    )
    
    # Server control group
    server_group = parser.add_mutually_exclusive_group()
    for action in ServerAction:
        server_group.add_argument(
            f'--{action.name.lower()}',
            action='store_true',
            help=f'{action.name.capitalize()} the Minecraft server'
        )
    
    # Update options
    parser.add_argument(
        '--auto-update',
        action='store_true',
        help='Run automated update process with player warnings'
    )
    
    # Configuration
    parser.add_argument(
        '--config',
        default=DEFAULT_CONFIG_PATH,
        help=f'Path to configuration file (default: {DEFAULT_CONFIG_PATH})'
    )
    
    return parser.parse_args()


def handle_server_action(manager: MinecraftModManager, action: ServerAction) -> int:
    """Handle server control actions."""
    match action:
        case ServerAction.STATUS:
            status = "running" if manager.verify_server_status() else "stopped"
            players = manager.get_player_count() if status == "running" else 0
            print(f"Server is {status} with {players} players online")
            return 0
            
        case ServerAction.START | ServerAction.STOP | ServerAction.RESTART:
            action_name = action.name.lower()
            print(f"{action_name.capitalize()}ing server...")
            if manager.control_server(action_name):
                print(f"Server {action_name}ed successfully")
                return 0
            print(f"Failed to {action_name} server")
            return 1


def main() -> int:
    """Command-line interface entry point."""
    try:
        args = parse_args()
        manager = MinecraftModManager(config_path=args.config)
        
        # Handle server actions
        for action in ServerAction:
            if getattr(args, action.name.lower()):
                return handle_server_action(manager, action)
        
        # Handle update commands
        if args.auto_update:
            print("Starting automated update process...")
            asyncio.run(manager.run_automated_update())
            return 0
        
        # Manual maintenance mode
        print("\nWarning: This will update all mods and restart the server.")
        print("Players will be warned with a countdown if any are online.")
        try:
            input("Press Enter to continue or Ctrl+C to cancel...")
            asyncio.run(manager.run_maintenance())
            return 0
        except KeyboardInterrupt:
            print("\nOperation cancelled by user")
            return 0

    except KeyboardInterrupt:
        print("\nOperation cancelled by user")
        return 0
    except Exception as e:
        print(f"Error: {str(e)}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main()) 