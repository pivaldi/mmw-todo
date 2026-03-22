import { timestampToDate, dateToTimestamp } from './timestamp.pipe';
import type { Timestamp } from '@bufbuild/protobuf/wkt';

describe('timestampToDate', () => {
  it('returns null for undefined', () => {
    expect(timestampToDate(undefined)).toBeNull();
  });

  it('converts a zero Timestamp to epoch', () => {
    const ts = { seconds: 0n, nanos: 0 } as Timestamp;
    expect(timestampToDate(ts)).toEqual(new Date(0));
  });

  it('converts seconds and nanos correctly', () => {
    // 2024-01-01T00:00:00.500Z → seconds=1704067200, nanos=500_000_000
    const ts = { seconds: 1704067200n, nanos: 500_000_000 } as Timestamp;
    expect(timestampToDate(ts)).toEqual(new Date(1704067200500));
  });
});

describe('dateToTimestamp', () => {
  it('converts epoch to zero Timestamp', () => {
    const ts = dateToTimestamp(new Date(0));
    expect(ts.seconds).toBe(0n);
    expect(ts.nanos).toBe(0);
  });

  it('round-trips through timestampToDate', () => {
    const original = new Date(1704067200500);
    const roundTripped = timestampToDate(dateToTimestamp(original));
    expect(roundTripped).toEqual(original);
  });
});
