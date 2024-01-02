# MGit
Implementing a git from scratch, referenced from [Write yourself a Git!](https://wyag.thb.lt/). To distinguish from `.git`, this tool employs `.mgit` as the repository for version control.

Prior to delving into the source code, it is suggested to familiarize yourself with the following related concepts.

## Concepts:

### Objects
- `object`. This is the form utilized by git for managing, organizing, and storing. For instance, blob, tree, commit, and tag are all considered objects.

    Objects generate a hash value through a hash algorithm to serve as the filename. The file content stores the specific content of various objects and it is placed under the `.git/objects` directory.

- `blob`. This stores the content of an individual file within the repository.
- `tree`. This depicts the content of the repository. A tree links blobs and subtrees together, organizing them into a structured tree, thus providing a basis for retrieving historical versions. `git checkout` is essentially finding the tree and restoring it to the system file based on its described organizational relationship.

### References
- Referencing Mechanism. In addition to storing objects via objects, git also utilizes a referencing mechanism for the convenience of organizing and managing them. References can be either directly or indirectly:
    - Those stored under `.git/refs` are direct references, meaning that their file content points directly to a certain object.
    - `.git/HEAD` generally stores indirect references, such as `ref: refs/heads/master`. The object hash value can be located and the specific file found through pointing to the file path.
- `tag`. A tag can be considered a reference pointing to any object. It can either be a lightweight tab or an informative tag, later being the tag object. Stored under the `refs/tags` directory.
- `branch`. A branch is essentially a reference pointing to a certain commit. Stored under `refs/heads` directory.
- `HEAD`. HEAD is a reference pointing to the current branch or a certain commit, aiding in locating your position in the historical version.

### Index File

- `index file` or `staging area`. To commit in Git, it is necessary to "stage" some changes using `git add` and `git rm` before committing them. This intermediate stage between the last and next commit is called the "staging area”.

    Git employs the `index file` to manage the staging area. The index file contains file information. Comparison with the tree pointed to by HEAD gives the differences between the staging area and HEAD; comparing this with the filesystem (actual files in the repository) gives the differences between the staging area and the filesystem; these two comparisons illustrate the `git status` command, indicating which files have been modified, added, or deleted.
    The `git add/rm` command is used to modify the index file content; then `git commit` is required to save the index file to the disk as a historical version.
- `commit`. The commit command mainly writes the flat structure in the index back to the disk according to the tree structure, links the commit and the tree together, adds some commit information, thus constituting a commit object. After writing the commit object to the disk, it also updates the commit pointed to by the branch.
