# Validate Interface

To ensure the Fastly Terraform Provider doesn't break the interface for users of the current release, we setup a real project using the current release and then compile the version of the provider from the `main` branch and run it against the project.
