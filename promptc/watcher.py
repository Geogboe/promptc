"""File watcher for auto-recompilation."""

import time
from pathlib import Path
from typing import Optional
from .compiler import PromptCompiler


class PromptWatcher:
    """Watches .prompt files and recompiles on changes."""

    def __init__(
        self,
        prompt_file: Path,
        target: str,
        output: Optional[Path] = None,
        interval: float = 1.0
    ):
        """
        Initialize watcher.

        Args:
            prompt_file: Path to .prompt file to watch
            target: Target format for compilation
            output: Optional output file path
            interval: Check interval in seconds (default: 1.0)
        """
        self.prompt_file = prompt_file
        self.target = target
        self.output = output
        self.interval = interval
        self.last_mtime = None

        if not prompt_file.exists():
            raise FileNotFoundError(f"Prompt file not found: {prompt_file}")

    def watch(self):
        """Start watching the file for changes."""
        print(f"Watching {self.prompt_file} for changes...")
        print(f"Target: {self.target}")
        if self.output:
            print(f"Output: {self.output}")
        print("Press Ctrl+C to stop")
        print()

        try:
            # Initial compilation
            self._compile()

            # Watch loop
            while True:
                time.sleep(self.interval)

                if self._file_changed():
                    print(f"\nDetected change in {self.prompt_file}")
                    self._compile()

        except KeyboardInterrupt:
            print("\n\nStopped watching")

    def _file_changed(self) -> bool:
        """Check if file has been modified."""
        if not self.prompt_file.exists():
            return False

        current_mtime = self.prompt_file.stat().st_mtime

        if self.last_mtime is None:
            self.last_mtime = current_mtime
            return False

        if current_mtime > self.last_mtime:
            self.last_mtime = current_mtime
            return True

        return False

    def _compile(self):
        """Compile the prompt file."""
        try:
            project_dir = self.prompt_file.parent
            compiler = PromptCompiler(project_dir=project_dir)

            result = compiler.compile(self.prompt_file, target=self.target)

            if self.output:
                self.output.parent.mkdir(parents=True, exist_ok=True)
                self.output.write_text(result, encoding="utf-8")
                print(f"✓ Compiled to {self.output}")
            else:
                print("✓ Compilation successful")
                print()
                print(result)

        except Exception as e:
            print(f"✗ Compilation failed: {e}")
