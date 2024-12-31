"""Main entry point for the Minecraft Mod Manager."""

import argparse
import asyncio
import logging
import sys

from .minecraft_mod_manager import MinecraftModManager
from .utils.constants import DEFAULT_CONFIG_PATH

def parse_args() -> argparse.Namespace:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="Minecraft Server Mod Manager",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    %(prog)s --config custom_config.toml --auto-update
    %(prog)s --maintenance
        """
    )
    
    parser.add_argument(
        "--config",
        type=str,
        default=DEFAULT_CONFIG_PATH,
        help="Path to configuration file (default: %(default)s)",
    )
    
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument(
        "--auto-update",
        action="store_true",
        help="Run automated update process",
    )
    group.add_argument(
        "--maintenance",
        action="store_true",
        help="Run manual maintenance process",
    )
    
    return parser.parse_args()

async def main() -> None:
    """Main entry point."""
    args = parse_args()
    
    try:
        manager = MinecraftModManager(args.config)
        
        if args.auto_update:
            await manager.run_automated_update()
        elif args.maintenance:
            await manager.run_maintenance()
            
    except Exception as e:
        logging.error(f"Error: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())