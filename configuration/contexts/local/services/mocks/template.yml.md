# Mock Services schema template

The mock services allows you to have a standard response from an endpoint, this is useful to for example mock a service we do not want to have running locally

```yaml
# this is the starting point to generate any mocking route, the reason is locally is folder agnostic
# so while we places all the mock services definitions in the same folder it does necessarily
# needs to be so, adding that allows locally to set the configuration in the right place
mockServices:
  # a mockService is an array, you can define multiple mocks per file and they just need to have
  # unique names, they also have a naming rule, it needs to be alphanumeric and can contain
  # dashes and underscores but it cannot contain spaces, for example example-definition
  - name: notifications
    # each mock can contain multiple routes to mock, each will need to have a name and that name
    # needs to be unique
    mockRoutes:
        # name of the mock route, this can contain letter, numbers, dashes and underscores
      - name: "connect"
        # the regex expression to match in the proxy to forward the mock service
        regex: ^\/api\/notification\/v1\/connect.*
        # building the response object, you can either have a complex object as a response in
        # json or just a simple string
        responds:
          # this will set the response content type header
          contentType: application/json
          # you either can have this, and it will respond a simple string, if the string is a
          # json object and the content type is set to json then it still looks like a json
          # response
          rawBody: "Ok"
          # or you can set a more complex body, this would return a json in that format
          body:
            enabled: false
            foo: "bar"
```
