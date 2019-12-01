-- Load URL paths from the file
function load_url_paths_from_file(file)
  lines = {}

  local f=io.open(file,"r")
  if f~=nil then
    io.close(f)
  else
    return lines
  end

  for line in io.lines(file) do
    if not (line == '') then
      lines[#lines + 1] = line
    end
  end

  return lines
end

-- Load URL paths from file
-- paths = load_url_paths_from_file("messages.txt")

-- print("multiplepaths: Found " .. #paths .. " paths")

-- Initialize the paths array iterator
counter = 1

request = function()
  -- Get the next paths array element
  -- url_path = paths[counter]
  url_path = "http://localhost:8080/chat/"

  counter = counter + 1

  -- If the counter is longer than the paths array length then reset it
  -- if counter > #paths then
  if counter > 20000 then
    counter = 1
  end

  -- Return the request object with the current URL path
  return wrk.format(nil, url_path .. tostring(counter))
end
