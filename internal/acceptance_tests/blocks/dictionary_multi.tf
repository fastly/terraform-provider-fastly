  dictionary {
    name = "{{.DICTIONARY_NAME_1}}"
  }

  dictionary {
    name          = "{{.DICTIONARY_NAME_2}}"
    write_only    = true
    force_destroy = true
  }
