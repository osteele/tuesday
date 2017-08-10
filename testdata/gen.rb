# Create a file data.csv in the current directory, containing format strings,
# and the reference date thus formatted.
#
# This is run manually rather than on go generate, so that go generate doesn't
# require a Ruby installation.

require 'CSV'
require 'time'

rt = Time.iso8601 "2006-01-02T15:04:05.123456789-05:00"
CSV.open(File.join(File.dirname(__FILE__), "data.csv"), "w") do |csv|
    for mod in ['', '-', '_', '^', '#'] do
        for c in ('A'..'Z').to_a + ('a'..'z').to_a do
            fmt = "%#{mod}#{c}"
            out = rt.strftime(fmt)
            next if out == fmt
            next if mod != '' && out == rt.strftime("%#{c}")
            csv << [fmt, out]
        end
    end
end
