-- Time-stamp: <2026-09-07 10:48:48 krylon>
--
-- /home/krylon/go/src/github.com/blicero/jazz/jes/testdata/test01.lua
-- created on 05. 09. 2026
-- (c) 2026 Benjamin Walkenhorst
--
-- Redistribution and use in source and binary forms, with or without
-- modification, are permitted provided that the following conditions
-- are met:
-- 1. Redistributions of source code must retain the copyright
--    notice, this list of conditions and the following disclaimer.
-- 2. Redistributions in binary form must reproduce the above copyright
--    notice, this list of conditions and the following disclaimer in the
--    documentation and/or other materials provided with the distribution.
--
-- THIS SOFTWARE IS PROVIDED BY BENJAMIN WALKENHORST "AS IS" AND
-- ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
-- IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
-- ARE DISCLAIMED.  IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS BE LIABLE
-- FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
-- DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
-- OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
-- HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
-- LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
-- OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
-- SUCH DAMAGE.


local job = {
   name = "test01",
   workdir = "/tmp",
   nice = 5,
   scheduledstart = "+120m",
   deadline = "+360m",
   env = {
      PATH = { "/bin", "/usr/bin", "/usr/local/bin" },
      LC_ALL = "en_US.UTF-8"
   },
   steps = {
      {
         command = "rm -r go-build* ts-out.*"
      },
      {
         command = "journalctl --vacuum-time=7days"
      },
      {
         command = {
            "sqlite3",
            "~/.newsroom/newsroom.db",
            "PRAGMA wal_checkpoint(full); VACUUM; ANALYZE;"
         }
      }
   }
}

return job
