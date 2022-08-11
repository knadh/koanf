### How to pass this test

This test can be used with any cloud storage,
which is available in RClone.

You must have installed RClone on your computer.

### Example: how to run this test with Google Drive

1. Create Google Drive account.
2. Create rclone configuration using this instruction:

https://rclone.org/drive/

Let's presume that the name of your remote storage, which
you put during the first step, is godrive1.

```
name> godrive1
```

3. Put the file koanf/mock/mock.json to the root directory
of your Google Drive storage.
4. Make sure that your configuration works. For example,
run this command:

```
rclone cat godrive1:mock.json
```

godrive1 is the name of your remote storage:
(2. name> godrive1), mock.json is in the root folder.

- you will see the content of the file in your console.

If your cloud storage contains buckets (Amazon S3,
minio, Backblaze B2) and the file mock.json is put into
the bucket1, the access to the file is:

```go
f := rclone.Provider(rclone.Config{Remote: remote, File: "bucket1/mock.json"})
```

5. Put the name of your remote storage in the file
'cloud.txt' without any special symbols. This file
should contain only this name. You can use godrive1
or godrive1: with the colon or without it.

6. Run:

```
go run main.go
```
