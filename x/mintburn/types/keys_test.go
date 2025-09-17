package types

import "testing"

func TestNormalizeChainIDBase(t *testing.T) {
    cases := []struct{
        in   string
        want string
    }{
        {"maanydex-1", "maanydex"},
        {"maanydex-12", "maanydex"},
        {"maanydex", "maanydex"},
        {"chain-test", "chain-test"},
        {"foo-abc", "foo-abc"},
        {"-1", "-1"},
        {"a-", "a-"},
        {"foo-01", "foo"},
        {"foo-0a", "foo-0a"},
        {"", ""},
    }
    for _, tc := range cases {
        got := NormalizeChainIDBase(tc.in)
        if got != tc.want {
            t.Fatalf("NormalizeChainIDBase(%q) = %q; want %q", tc.in, got, tc.want)
        }
    }
}

func TestMintburnEscrowChannelIDFor(t *testing.T) {
    cases := []struct{
        in   string
        want string
    }{
        {"maanydex-3", "mintburn-maanydex"},
        {"abc", "mintburn-abc"},
    }
    for _, tc := range cases {
        got := MintburnEscrowChannelIDFor(tc.in)
        if got != tc.want {
            t.Fatalf("MintburnEscrowChannelIDFor(%q) = %q; want %q", tc.in, got, tc.want)
        }
    }
}

