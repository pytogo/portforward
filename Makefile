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

lint: ## check style with flake8
	flake8 portforward.py --count --select=E9,F63,F7,F82 --show-source --statistics
	flake8 portforward.py --count --exit-zero --max-complexity=10 --max-line-length=88 --statistics

test: ## run tests quickly with the default Python
	pytest

test-all: ## run tests on every Python version with tox
	tox

coverage: ## check code coverage quickly with the default Python
	coverage run --source portforward -m pytest
	coverage report -m
	coverage html
	$(BROWSER) htmlcov/index.html

docs-check:
	python setup.py check -r -s

docs: ## generate Sphinx HTML documentation, including API docs
	rm -f docs/portforward.rst
	rm -f docs/modules.rst
	sphinx-apidoc -o docs/ . tests setup.py
	$(MAKE) -C docs clean
	$(MAKE) -C docs html
	$(BROWSER) docs/_build/html/index.html

servedocs: docs ## compile the docs watching for changes
	watchmedo shell-command -p '*.rst' -c '$(MAKE) -C docs html' -R -D .

install: clean ## install the package to the active Python's site-packages
	python setup.py install

# ===== LINUX =====
release-test-linux: clean ## package and upload a release to test.pypi
	setuptools-golang-build-manylinux-wheels --golang 1.16.6
	python setup.py sdist
	twine upload --repository testpypi dist/*

release-linux: clean ## package and upload a release for linux
	setuptools-golang-build-manylinux-wheels --golang 1.16.6
	python setup.py sdist
	twine upload dist/*

# ===== WINDOWS =====
release-test-windows: ## package and upload a release to test.pypi
	python39 setup.py bdist_wheel
	python38 setup.py bdist_wheel
	python37 setup.py bdist_wheel
	python36 setup.py bdist_wheel
	python39 -m twine upload --repository testpypi dist\*

release-windows: ## package and upload a release for Linux
	python39 setup.py bdist_wheel
	python38 setup.py bdist_wheel
	python37 setup.py bdist_wheel
	python36 setup.py bdist_wheel
	python39 -m twine upload dist\*

# ===== MACOS =====
release-test-macos: clean ## package and upload a release to test.pypi
	python3.9 setup.py bdist_wheel
	python3.8 setup.py bdist_wheel
	python3.7 setup.py bdist_wheel
	python3.6 setup.py bdist_wheel
	python3.9 -m twine upload --repository testpypi dist/*

release-macos: clean ## package and upload a release for MacOS
	python3.9 setup.py bdist_wheel
	python3.8 setup.py bdist_wheel
	python3.7 setup.py bdist_wheel
	python3.6 setup.py bdist_wheel
	python3.9 -m twine upload dist/*
