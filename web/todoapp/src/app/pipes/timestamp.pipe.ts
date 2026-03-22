import { Pipe, PipeTransform } from '@angular/core';
import type { Timestamp } from '@bufbuild/protobuf/wkt';

export function timestampToDate(ts: Timestamp | undefined): Date | null {
  if (!ts) return null;
  return new Date(Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000));
}

export function dateToTimestamp(d: Date): Timestamp {
  const ms = d.getTime();
  return {
    seconds: BigInt(Math.floor(ms / 1000)),
    nanos: (ms % 1000) * 1_000_000,
  } as Timestamp;
}

@Pipe({ name: 'timestamp' })
export class TimestampPipe implements PipeTransform {
  transform(ts: Timestamp | undefined): Date | null {
    return timestampToDate(ts);
  }
}
