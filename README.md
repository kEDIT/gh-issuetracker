# gh-issuetracker

---

### Description

This is toy web application, which fetches issues from a public GitHub repository and renders a summary.

### Usage
* #### Using Docker:
    * ```docker pull kedit/gh-issuetracker```
    * ```docker run -p 8090:8090 kedit/gh-issuetracker```
    * Point your browser to ```localhost:8090```
* #### Run from source:
    * ```git clone https://github.com/kedit/gh-issuetracker .```
    * ```cd gh-issuetracker```
    * ```go run main.go```

### Dependencies

* I used a client library for handling API requests to save time, instead of manually defining the response object and deserializing etc. It is linked [here](https://github.com/google/go-github).

* HTMX 1.9.5 for front end interactivity
* Bootstrap 5 for styling