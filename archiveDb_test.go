package mclib

import (
  "fmt"
  "time"
  "testing"
  "math/rand"
  // "github.com/aws/aws-sdk-go/service/s3"
  "github.com/stretchr/testify/assert"
)


func TestArchiveStringNoObject (t *testing.T) {
  s := testServer(t, false)
  a := NewArchive(ServerSnapshot, s.User, s.Name, s.ArchiveBucket, nil)
  assert.Contains(t, a.String(), s.User)
  assert.Contains(t, a.String(), "NO-REFERENCE-TO-S3-OBJECT")
}

func TestArchiveStringWithObject (t *testing.T) {
  s := testServer(t, false)
  path, _ := s.archivePath(ServerSnapshot)
  o := testS3Object(path)
  a := NewArchive(ServerSnapshot, s.User, s.Name, s.ArchiveBucket, o)
  assert.Contains(t, a.String(), path)
  assert.NotContains(t, a.String(), "NO-REFERENCE-TO-S3-OBJECT")
  assert.NotNil(t, a.S3Object)
}

// Spec: /<user>/<server>/<archive-type-string>
func TestArchivePath(t *testing.T) {
  // u := randomUniqueUserNames(1)[0]
  u := randUserName()
  // tm := time.Time.Now()
  s := fmt.Sprintf("%s-%s", u, randServerName())
  tp := typeToPathElement[ServerSnapshot]

  ex := S3PathJoin(u,s,tp)
  assert.Equal(t, u + S3Delim + s + S3Delim + tp + S3Delim, ex, "S3PathJoin is not working properly.")

  ap := archivePath(u,s,ServerSnapshot)
  assert.Equal(t, ap, ex, "ArchivePath didn't work.")
  // fmt.Printf("Snapshot ArchivePath: %s\n", ap)
}

// Full path Spec
// http://s3.amazonaws.com/<bucket>/<user>/<server>/<archive-type-string>/<RFC3339-time>-<user>-server-<archive-ext>
func TestSnapshotPath(t *testing.T) {
  bucket := DefaultTestBucket
  user := randUserName()
  server := fmt.Sprintf("%s", randServerName()) 
  aType := ServerSnapshot
  aTypeES := typeToPathElement[aType]
  when := time.Now()
  whenString := time.Now().Format(time.RFC3339)
  archiveExt := typeToFileExt[ServerSnapshot]

  // File Name
  exFileName := whenString + "-" + user + "-" + server + archiveExt
  acFileName := archiveFileName(user, server, when, ServerSnapshot)
  assert.Equal(t, exFileName, acFileName, "archiveFileName isn't right")
  // fmt.Printf("Snapshot File Name: %s\n", acFileName)

  // Full URI
  exURI := S3BaseURI + S3Delim + bucket + S3Delim + user + S3Delim + 
    server + S3Delim + aTypeES + S3Delim + exFileName
  acURI := ServerSnapshotURI(bucket, user, server, exFileName)
  assert.Equal(t, exURI, acURI, "URI isn't right.")

  // fmt.Printf("%s\n", exURI)
  // fmt.Printf("%s\n", acURI)
}

func getRandomArchiveType() (ArchiveType) {
  return AllArchiveTypes[rand.Int() % len(AllArchiveTypes)]
}

func TestArchiveType(t *testing.T) {
  for ty, ts := range archiveTypeToString {
    assert.Equal(t, ty, archiveStringToType[ts], "Failed to match %s, %s\n",ty.String(), archiveStringToType[ts].String())
  }
}

func TestGetSnapshots (t *testing.T) {
  // Config
  iters := 100
  numUniqueUsers := 10

  snapCount := 0
  userSet := make(map[string]bool)
  users := randomUniqueUserNames(numUniqueUsers)
  user := users[rand.Int() % numUniqueUsers]
  archiveMap := make(ArchiveMap, iters)
  for i := 1; i < iters; i++ {

    // Get a user
    un := users[rand.Int() % len(users)]
    _, ok := userSet[un]
    if !ok {userSet[un] = true} // track how many unique users we added to the map.

    // and a server
    var sn string
    if rand.Int() % 2 == 0 {
      sn = fmt.Sprintf("%s-TestServer-1", un )
    } else {
      sn = fmt.Sprintf("%s-TestServer-2", un)
    }
    s := testServer(t, true)
    s.Name = sn

    // Pick an archive type
    t := getRandomArchiveType()
    if t == ServerSnapshot && un == user {snapCount++}
    path, _ := s.archivePath(ServerSnapshot)

    // Add the archive to the map.
    a := NewArchive(t, un, sn, "test-bucket", testS3Object(path))
    archiveMap.Add(a)
  }
  assert.Len(t, archiveMap, len(userSet), "ArchiveMap didn't add the right number of archives.")
  snaps := archiveMap.GetArchives(user, ServerSnapshot)
  assert.Len(t, snaps, snapCount, "Didn't get the right number of snapshots back.")
}

