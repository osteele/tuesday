# Create a file data.csv in the current directory, containing format strings,
# and the reference date thus formatted.
#
# This is run manually rather than on go generate, so that go generate doesn't
# require a Ruby installation.

require 'CSV'
require 'time'

rt = Time.iso8601 "2006-01-02T15:04:05.123456789-05:00"
CSV.open(File.join(File.dirname(__FILE__), "data.csv"), "w") do |csv|
    for flag in ['', '-', '_', '^', '#', '0'] do
        for c in ('A'..'Z').to_a + ('a'..'z').to_a + %w[+ %] do
            fmt = "%#{flag}#{c}"
            out = rt.strftime(fmt)
            # Skip ignored conversions. Some of these actually do something applied to
            # Date, so we don't want to assert that they do nothing.
            next if out == fmt
            # did the flag make a difference?
            next if flag != '' && out == rt.strftime("%#{c}")
            csv << [fmt, out]
        end
    end
end
