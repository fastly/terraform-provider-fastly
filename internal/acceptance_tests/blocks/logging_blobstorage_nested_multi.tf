logging_blobstorage {
  name      = "{{.LOGGING_BLOBSTORAGE_NAME_1}}"
  container = "{{.CONTAINER_NAME}}"
  authentication = {
    account_name = "teststorageaccount"
    sas_token    = "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"
  }
}

logging_blobstorage {
  name      = "{{.LOGGING_BLOBSTORAGE_NAME_2}}"
  container = "{{.CONTAINER_NAME}}"
  authentication = {
    account_name = "teststorageaccount"
    sas_token    = "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"
  }
}
