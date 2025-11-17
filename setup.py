"""Setup configuration for promptc."""

from setuptools import setup, find_packages
from pathlib import Path

# Read the README for long description
this_directory = Path(__file__).parent
long_description = (this_directory / "README.md").read_text(encoding="utf-8") if (this_directory / "README.md").exists() else ""

setup(
    name="promptc",
    version="0.1.0",
    author="Your Name",
    author_email="your.email@example.com",
    description="A prompt compiler for agentic programming - manage and compile LLM instructions into different formats",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/yourusername/promptc",
    packages=find_packages(),
    include_package_data=True,
    package_data={
        "promptc": [
            "defaults/patterns/*.prompt",
            "defaults/constraints/*.prompt",
        ],
    },
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "Topic :: Software Development :: Code Generators",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
    python_requires=">=3.8",
    install_requires=[
        "PyYAML>=6.0",
    ],
    entry_points={
        "console_scripts": [
            "promptc=promptc.cli:main",
        ],
    },
    keywords="prompt llm ai claude cursor aider compiler agentic",
)
