"""Setup script for Minecraft Mod Manager."""

from setuptools import find_namespace_packages, setup

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="minecraft-mod-manager",
    version="1.0.0",
    author="dacrab",
    author_email="dacrab@github.com",
    description="A comprehensive Minecraft server mod management tool",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/dacrab/minecraft-mod-manager",
    packages=find_namespace_packages(include=["minecraft_mod_manager*"]),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Environment :: Console",
        "Intended Audience :: System Administrators",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Games/Entertainment",
        "Topic :: System :: Software Distribution",
        "Topic :: Utilities",
    ],
    python_requires=">=3.9",
    install_requires=[
        "aiohttp>=3.9.1",
        "tqdm>=4.66.1",
        "toml>=0.10.2",
    ],
    entry_points={
        "console_scripts": [
            "minecraft-mod-manager=minecraft_mod_manager.minecraft_mod_manager:main",
        ],
    },
    include_package_data=True,
    package_data={
        "minecraft_mod_manager": ["config/*.toml", "config/config.toml"],
    },
)
