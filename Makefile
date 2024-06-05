.PHONY: clean clean-test clean-pyc clean-build docs help
.DEFAULT_GOAL := help

define BROWSER_PYSCRIPT
import os, webbrowser, sys

from urllib.request import pathname2url

webbrowser.open("file://" + pathname2url(os.path.abspath(sys.argv[1])))
endef
export BROWSER_PYSCRIPT

define PRINT_HELP_PYSCRIPT
import re, sys

for line in sys.stdin:
	match = re.match(r'^([a-zA-Z_-]+):.*?## (.*)$$', line)
	if match:
		target, help = match.groups()
		print("%-20s %s" % (target, help))
endef
export PRINT_HELP_PYSCRIPT

BROWSER := python -c "$$BROWSER_PYSCRIPT"

help:
	@python -c "$$PRINT_HELP_PYSCRIPT" < $(MAKEFILE_LIST)

clean: clean-build clean-pyc clean-test ## remove all build, test, coverage and Python artifacts

clean-build: ## remove build artifacts
	rm -fr build/
	rm -fr dist/
	rm -fr .eggs/
	find . -name '*.egg-info' -exec rm -rf {} +
	find . -name '*.egg' -exec rm -rf {} +

clean-pyc: ## remove Python file artifacts
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +
	find . -name '*~' -exec rm -f {} +
	find . -name '__pycache__' -exec rm -fr {} +

clean-test: ## remove test and coverage artifacts
	rm -fr .tox/
	rm -f .coverage
	rm -fr htmlcov/
	rm -fr .pytest_cache

check: ## check style with flake8 and black - also check types
	black --check python
	mypy python
	flake8 python --count --select=E9,F63,F7,F82 --show-source --statistics
	flake8 python --count --exit-zero --max-complexity=10 --max-line-length=88 --statistics

docs: ## generate Sphinx HTML documentation, including API docs
	rm -f docs/portforward.rst
	rm -f docs/modules.rst
	sphinx-apidoc -o docs/ . tests
	$(MAKE) -C docs clean
	$(MAKE) -C docs html
	$(BROWSER) docs/_build/html/index.html

servedocs: docs ## compile the docs watching for changes
	watchmedo shell-command -p '*.rst' -c '$(MAKE) -C docs html' -R -D .

release-linux: clean ## creates and release linux wheels
	mkdir -p dist
	docker run --rm -v $(PWD):/io ghcr.io/pyo3/maturin build --release -i python3.8 --out dist --strip
	docker run --rm -v $(PWD):/io ghcr.io/pyo3/maturin build --release -i python3.9 --out dist --strip
	docker run --rm -v $(PWD):/io ghcr.io/pyo3/maturin build --release -i python3.10 --out dist --strip
	docker run --rm -v $(PWD):/io ghcr.io/pyo3/maturin build --release -i python3.11 --out dist --strip
	docker run --rm -v $(PWD):/io ghcr.io/pyo3/maturin build --release -i python3.12 --out dist --strip

	maturin sdist --out dist

	twine upload dist/*

release-macos: clean ## creates and release macos wheels
	maturin build --release --target aarch64-apple-darwin --zig -i python3.8 --out dist --strip
	maturin build --release --target aarch64-apple-darwin --zig -i python3.9 --out dist --strip
	maturin build --release --target aarch64-apple-darwin --zig -i python3.10 --out dist --strip
	maturin build --release --target aarch64-apple-darwin --zig -i python3.11 --out dist --strip
	maturin build --release --target aarch64-apple-darwin --zig -i python3.12 --out dist --strip

	maturin build --release -i python3.8 --out dist --strip
	maturin build --release -i python3.9 --out dist --strip
	maturin build --release -i python3.10 --out dist --strip
	maturin build --release -i python3.11 --out dist --strip
	maturin build --release -i python3.12 --out dist --strip

	twine upload dist/*

release-windows: clean ## creates and release window wheels
	maturin build --release -i python38 --out dist --strip
	maturin build --release -i python39 --out dist --strip
	maturin build --release -i python310 --out dist --strip
	maturin build --release -i python311 --out dist --strip
	maturin build --release -i python312 --out dist --strip

	twine upload dist/*
