#!/usr/bin/env python

"""The setup script."""

from setuptools import setup, Extension

with open("README.rst") as readme_file:
    readme = readme_file.read()

with open("HISTORY.rst") as history_file:
    history = history_file.read()

requirements = []

test_requirements = ["pytest>=3", ]

setup(
    author="Sebastian Ziemann",
    author_email="corka149@mailbox.org",
    python_requires=">=3.6",
    classifiers=[
        "Development Status :: 2 - Pre-Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Natural Language :: English",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.6",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
    ],
    description="Kubernetes Port-Forward Go-Edition For Python ",
    install_requires=requirements,
    license="MIT license",
    long_description=readme + "\n\n" + history,
    include_package_data=True,
    keywords="portforward",
    name="portforward",
    py_modules=["portforward"],
    test_suite="tests",
    tests_require=test_requirements,
    url="https://github.com/pytogo/portforward",
    version="0.2.8",
    zip_safe=False,
    # Go part
    setup_requires=['setuptools-golang'],
    build_golang={'root': 'github.com/pytogo/portforward'},
    ext_modules=[
        Extension(
            "_portforward", ["main.go"],
            py_limited_api=True, define_macros=[('Py_LIMITED_API', None)],
        )
    ]
)
