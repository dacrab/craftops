"""Command-line interface for Minecraft Mod Manager."""

import argparse
import asyncio
import sys
import textwrap

from .minecraft_mod_manager import MinecraftModManager
from .utils.constants import DEFAULT_CONFIG_PATH

def parse_args() -> argparse.Namespace:
    """Parse and validate command line arguments."""
    parser = argparse.ArgumentParser(
        description='Minecraft Mod Manager - Server management and mod update tool',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=textwrap.dedent("""
            Examples:
              %(prog)s --status
              %(prog)s --start
              %(prog)s --auto-update
              %(prog)s --config custom_config.jsonc --auto-update
        """)
    )
    
    # Server control group
    server_group = parser.add_mutually_exclusive_group()
    server_group.add_argument(
        '--start',
        action='store_true',
        help='Start the Minecraft server'
    )
    server_group.add_argument(
        '--stop',
        action='store_true',
        help='Stop the Minecraft server'
    )
    server_group.add_argument(
        '--restart',
        action='store_true',
        help='Restart the Minecraft server'
    )
    server_group.add_argument(
        '--status',
        action='store_true',
        help='Check server status and player count'
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

def main() -> int:
    """
    Command-line interface entry point.
    Supports:
    - Auto-update
    - Server start/stop/restart
    - Status check
    - Manual maintenance
    
    Returns:
        int: Exit code (0 for success, 1 for error)
    """
    try:
        args = parse_args()
        
        # Initialize manager
        manager = MinecraftModManager(config_path=args.config)
        
        # Handle server status check
        if args.status:
            status = "running" if manager.verify_server_status() else "stopped"
            players = manager.get_player_count() if status == "running" else 0
            print(f"Server is {status} with {players} players online")
            return 0
        
        # Handle server control commands
        for action in ['start', 'stop', 'restart']:
            if getattr(args, action):
                print(f"{action.capitalize()}ing server...")
                if manager.control_server(action):
                    print(f"Server {action}ed successfully")
                    return 0
                else:
                    print(f"Failed to {action} server")
                    return 1
        
        # Handle update commands
        if args.auto_update:
            print("Starting automated update process...")
            asyncio.run(manager.run_automated_update())
            return 0
            
        else:
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
    main() 