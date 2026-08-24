package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

// newTestClient lance un miniredis en mémoire et retourne un *Client connecté dessus.
func newTestClient(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	client, err := NewClient("redis://" + mr.Addr())
	require.NoError(t, err, "NewClient doit se connecter à miniredis")

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client, mr
}

func sampleJob(id string) TrackJob {
	return TrackJob{
		TrackID:  id,
		FilePath: "/data/audio/" + id + ".mp3",
		Title:    "Track " + id,
		Artist:   "Test Artist",
		Duration: 180,
	}
}

// Une URL invalide doit renvoyer une erreur.
func TestNewClient_InvalidURL(t *testing.T) {
	_, err := NewClient("pas-une-url-redis")
	require.Error(t, err)
}

// Push ajoute un élément et Len le compte.
func TestPush_IncrementsLen(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()
	const streamID = "stream-123"

	n, err := client.Len(ctx, streamID)
	require.NoError(t, err)
	require.Equal(t, int64(0), n)

	require.NoError(t, client.Push(ctx, streamID, sampleJob("t1")))

	n, err = client.Len(ctx, streamID)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
}

// La clé Redis doit être "queue:stream:<streamID>".
func TestPush_UsesCorrectKey(t *testing.T) {
	client, mr := newTestClient(t)
	ctx := context.Background()

	require.NoError(t, client.Push(ctx, "abc", sampleJob("t1")))
	require.True(t, mr.Exists("queue:stream:abc"))
}

// Une track poussée doit ressortir identique (intégrité du round-trip JSON).
func TestPushPop_RoundTrip(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	in := sampleJob("t42")
	require.NoError(t, client.Push(ctx, "stream-rt", in))

	out, err := client.Pop(ctx, "stream-rt")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, in, *out)
}

// LPUSH + BRPOP => ordre FIFO.
func TestPushPop_FIFOOrder(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()
	const streamID = "stream-fifo"

	require.NoError(t, client.Push(ctx, streamID, sampleJob("t1")))
	require.NoError(t, client.Push(ctx, streamID, sampleJob("t2")))
	require.NoError(t, client.Push(ctx, streamID, sampleJob("t3")))

	for _, want := range []string{"t1", "t2", "t3"} {
		out, err := client.Pop(ctx, streamID)
		require.NoError(t, err)
		require.NotNil(t, out)
		require.Equal(t, want, out.TrackID)
	}
}

// File vide => (nil, nil) après le blockTimeout, pas une erreur.
func TestPop_EmptyReturnsNil(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	start := time.Now()
	out, err := client.Pop(ctx, "stream-vide")
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.Nil(t, out)
	require.GreaterOrEqual(t, elapsed, 1500*time.Millisecond)
}

// Clear vide la file.
func TestClear_EmptiesQueue(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()
	const streamID = "stream-clear"

	require.NoError(t, client.Push(ctx, streamID, sampleJob("t1")))
	require.NoError(t, client.Push(ctx, streamID, sampleJob("t2")))

	require.NoError(t, client.Clear(ctx, streamID))

	n, err := client.Len(ctx, streamID)
	require.NoError(t, err)
	require.Equal(t, int64(0), n)
}

// Deux streams = deux files indépendantes.
func TestQueues_AreIsolatedPerStream(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	require.NoError(t, client.Push(ctx, "stream-A", sampleJob("a1")))
	require.NoError(t, client.Push(ctx, "stream-B", sampleJob("b1")))

	outA, err := client.Pop(ctx, "stream-A")
	require.NoError(t, err)
	require.NotNil(t, outA)
	require.Equal(t, "a1", outA.TrackID)

	nB, err := client.Len(ctx, "stream-B")
	require.NoError(t, err)
	require.Equal(t, int64(1), nB)
}
